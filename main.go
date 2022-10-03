package main

import (
	"github.com/copo888/copo_schedule/cronjob"
	"github.com/neccoys/promx"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
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

	var logConf logx.LogConf
	logConf.Mode = "file"
	logConf.Level = "info"
	logConf.KeepDays = 30
	logConf.Path = "logs"
	logx.MustSetup(logConf)

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
	//c.AddJob("*/90 * * * * ?",
	//	cron.NewChain().
	//		Then(&cronjob.ProxyToChannel{}),
	//)

	/**
	 * 处里回调发生还款失败异及等待还款的提单，重新补还款机制(还款失败，代表回调成功，但还款在写入资料库时异常)
	 * 备注：每3分钟处理一次还款
	 */
	c.AddJob("0 0/3 * * * * ?", //3分鐘
		cron.NewChain().
			Then(&cronjob.HandleRepayment{}),
	)

	/**
	 * 处里超过5分钟渠道尚未回调的代付提单
	 * (备注：主动前往渠道查询结果，如果已确认为成功或失败，则透过回调机制写回diorpayment)
	 * 备注：5分钟处理一次还款
	 */
	//c.AddJob("0 0/5 * * * * ?", //5分鐘
	//	cron.NewChain().
	//		Then(&cronjob.QueryTransaction{}),
	//)

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

	// (計算月收益報表Schedule) 每月5號 03:00:00執行 '
	//c.AddJob("0 0 3 5 * ?",
	c.AddJob("0 0/10 * * * ?",
		cron.NewChain().
			Then(&cronjob.MonthProfitReport{}),
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
