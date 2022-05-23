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

	//c.AddJob("*/5 * * * * ?",
	//	cron.NewChain().
	//		Then(&cronjob.ProxyToChannel{}),
	//)

	// (補算傭金利潤Schedule) 整點開始每5分鐘執行
	c.AddJob("0 0/5 * * * ?",
		cron.NewChain().
			Then(&cronjob.CalculateProfit{}),
	)

	// (計算月傭金報表Schedule) 每月2號 00:00:00執行
	c.AddJob("0 0 0 2 * ?",
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
