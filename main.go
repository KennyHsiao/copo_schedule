package main

import (
	"github.com/copo888/copo_schedule/cronjob"
	"github.com/neccoys/promx"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func init() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("etc")

	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(0)
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU() - 2)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	logger := cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))

	c := cron.New(
		cron.WithLogger(logger),
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)),
	)

	//c.AddJob("* * * * * *",
	//	cron.NewChain().
	//		Then(&cronjob.TestJob{}),
	//)

	//90 秒抓取代付單[代處理]重送渠道
	c.AddJob("*/90 * * * * ?",
		cron.NewChain().
			Then(&cronjob.ProxyToChannel{}),
	)

	//排程每3分钟取出代付提单的还款状态为[3：还款失败][1:待还款][不等于人工处里]的提单，进行还款处理
	//3.1.还款前，前往渠道查询提单的目前状态，并依据下面查询到的规则做处理
	//(1).成功提单：指交易成功(已完成代付)，将提单转为成功提单，并执行结单。
	//(2).失败提单：指无此提单号或交易失败...等相关交易异常，将提单直接还款并结单。
	//(3).待处理及处理中提单：将提单转为已上传及处理中，等待回调。
	//(4).无此查询通道或其他查询异常：将提单转为人工处里，由后台管理人员处理提单还款或转成功。
	c.AddJob("0 0/3 * * * * ?", //3分鐘
		cron.NewChain().
			Then(&cronjob.ToPersonHandling{}),
	)

	// (補算傭金利潤Schedule) 整點開始每5分鐘執行
	c.AddJob("0 0/5 * * * ?",
		cron.NewChain().
			Then(&cronjob.CalculateProfit{}),
	)

	// (計算月傭金報表Schedule) 每月2號 03:00:00執行
	c.AddJob("0 0 3 2 * ?",
		cron.NewChain().
			Then(&cronjob.CommissionMonthReport{}),
	)

	c.Start()

	// prometheus
	promx.NewServe(
		viper.GetString("PROMETHEUS_NAME"),
		viper.GetString("PROMETHEUS_PATH"),
		viper.GetString("PROMETHEUS_PORT"),
	).Start()

	defer func() {
		c.Stop()
		log.Println("Copo Schedule Shutdown!")
	}()
}
