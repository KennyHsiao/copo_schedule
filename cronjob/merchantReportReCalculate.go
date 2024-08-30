package cronjob

import (
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"runtime"
	"sync"
	"time"
)

type ReCalMerchantReport struct {
	logx.Logger
	ctx context.Context
}

/*
* 確保close(jobQueue) -
 */
func (l *ReCalMerchantReport) Run() {
	url := fmt.Sprintf("%s:8080/api/v1/report/merchantreport/settlement", viper.GetString("SERVER"))

	datesList := generateDatesInterval(5)

	var wg sync.WaitGroup
	jobQueue := make(chan []string, len(datesList))
	//reQueue := make(chan []string, len(datesList))
	results := make(chan ReCalResult, len(datesList))
	span := trace.SpanFromContext(l.ctx)

	// 啟動worker
	for i := 0; i < len(datesList); i++ {
		wg.Add(1)
		go l.calculateWorker(url, &wg, results, jobQueue, &span)
	}

	// 派發任務
	for i := 0; i < len(datesList); i++ {
		jobQueue <- []string{datesList[i][0], datesList[i][1]}
	}

	// 處理重試邏輯
	//go func(reQueue <-chan []string, jobQueue chan<- []string) {
	//	for job := range reQueue {
	//		logx.WithContext(l.ctx).Infof("reQueue received job: %v\n", job)
	//		select {
	//		case jobQueue <- job: // 嘗試重新加入到jobQueue中
	//		default:
	//			// 如果jobQueue滿了，可以選擇等待一會再試，或者放棄此次重試
	//			logx.WithContext(l.ctx).Errorf("Job queue is full, skipping retry for job: %v\n", job)
	//		}
	//	}
	//}(reQueue, jobQueue)

	go func() {
		close(jobQueue)
		wg.Wait()      // 等待所有worker完成
		close(results) // 關閉results通道
	}()

	// 等待所有結果並打印
	for res := range results {
		logx.WithContext(l.ctx).Infof("Result for job [%s - %s]: Code: %s, Message: %s\n", res.Dates[0], res.Dates[1], res.Code, res.Message)
		fmt.Printf("Number of goroutines: %d\n", runtime.NumGoroutine())
	}

	logx.WithContext(l.ctx).Infof("All Work Done!\n")
}

func (l ReCalMerchantReport) calculateWorker(url string, wg *sync.WaitGroup, results chan<- ReCalResult, jobQueue <-chan []string, span *trace.Span) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobQueue:
			if !ok {
				// 如果通道關閉且無數據
				logx.WithContext(l.ctx).Infof("Job queue closed, worker exiting")
				return
			}
			var success bool
			for attempt := 0; attempt < 3; attempt++ {
				if success = l.tryJob(url, job, span, results); success {
					break
				}
				if attempt < 3 {
					fmt.Printf("Job [%s - %s] Try %d times\n", job[0], job[1], attempt)
					continue
				} else {
					results <- ReCalResult{Dates: job, Code: "MaxRetry", Message: "Max retry attempts reached"}
				}
			}
		}
	}
	logx.WithContext(l.ctx).Infof("Worker finished processing jobs\n")
}

func (l ReCalMerchantReport) tryJob(url string, job []string, span *trace.Span, results chan<- ReCalResult) bool {
	res, err := gozzle.Post(url).Timeout(20).Trace(*span).JSON(MerchantReportSettlementRequest{StartAt: job[0], EndAt: job[1]})

	//模擬返回錯誤
	//if strings.EqualFold(fmt.Sprintf("%s - %s", job[0], job[1]), "2024-07-01 00:00:00 - 2024-07-05 23:59:59"){
	//	logx.WithContext(l.ctx).Errorf("Job [%s - %s] failed: %s\n", job[0], job[1], "模擬錯誤")
	//	return false
	//}

	if err != nil {
		logx.WithContext(l.ctx).Errorf("Job [%s - %s] failed: %+v\n", job[0], job[1], err)
		return false
	}

	if res.Status() != 200 {
		logx.WithContext(l.ctx).Errorf("Job [%s - %s] HTTP error: %d\n", job[0], job[1], res.Status())
		return false
	}

	response := &CalculateResponse{}
	if err := res.DecodeJSON(response); err != nil {
		logx.WithContext(l.ctx).Errorf("Failed to decode response for job [%s - %s]: %v\n", job[0], job[1], err)
		return false
	}

	if response.Code != "0" {
		logx.WithContext(l.ctx).Errorf("Job [%s - %s] business logic failed: Code: %s, Message: %s\n", job[0], job[1], response.Code, response.Message)
		return false
	}

	// 如果成功，將結果發送到results通道
	results <- ReCalResult{Dates: job, Code: response.Code, Message: response.Message}
	return true
}

func generateDatesInterval(intervalDays int) [][]string {
	now := time.Now()

	// 调整到上上个月
	previousMonth := now.AddDate(0, -1, 0)

	// 获取上上个月的第一天
	firstDayOfPreviousMonth := time.Date(previousMonth.Year(), previousMonth.Month(), 1, 0, 0, 0, 0, time.Local)

	// 获取上上个月的最后一天
	lastDayOfPreviousMonth := firstDayOfPreviousMonth.AddDate(0, 1, -1)

	// 初始化结果切片
	var datesList [][]string

	// 从上上个月的第一天开始，每次增加5天，直到超过上上个月的最后一天
	for current := firstDayOfPreviousMonth; !current.After(lastDayOfPreviousMonth); {
		var internaDate []string
		if current.AddDate(0, 0, intervalDays).After(lastDayOfPreviousMonth) {
			// 先记录当前日期
			internaDate = append(internaDate, current.Format("2006-01-02")+" 00:00:00")
			// 再记录月末日期，并终止循环
			internaDate = append(internaDate, lastDayOfPreviousMonth.Format("2006-01-02")+" 23:59:59")
			datesList = append(datesList, internaDate)
			break
		} else {
			internaDate = append(internaDate, current.Format("2006-01-02")+" 00:00:00")
			// 正常增加5天并记录日期
			current = current.AddDate(0, 0, intervalDays)
			internaDate = append(internaDate, current.AddDate(0, 0, -1).Format("2006-01-02")+" 23:59:59")
		}
		datesList = append(datesList, internaDate)
	}

	return datesList
}

type MerchantReportSettlementRequest struct {
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
}

type ReCalResult struct {
	Dates   []string `json:"dates"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
}

type CalculateResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
