package main

import (
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

	//c.AddJob("*/5 * * * * ?",
	//	cron.NewChain().
	//		Then(&cronjob.CalculateProfit{}),
	//)

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
