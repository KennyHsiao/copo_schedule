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

	//transactionClient := zrpc.MustNewClient(zrpc.RpcClientConf{
	//	Target: fmt.Sprintf("consul://%s/transaction.rpc?wait=14s", viper.GetString("CONSUL_TARGET")),
	//})
	//transactionRpc := transactionclient.NewTransaction(transactionClient)

	c := cron.New(
		cron.WithLogger(logger),
		cron.WithSeconds(),
		//cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)),
	)

	//c.AddJob("* * * * * *",
	//	cron.NewChain().
	//		Then(&cronjob.TestJob{}),
	//)

	//90秒抓取代付單[代處理]反查渠道
	c.AddJob("0 0/1 * * * ?",
		cron.NewChain(cron.SkipIfStillRunning(logger)).
			Then(&cronjob.ProxyToChannel{}),
	)

	//1分鐘查餘額
	//c.AddJob("0 0/1 * * * * ?", //1分鐘)
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.QueryChannelBalance{}),
	//)
	/**
	 * 处里回调发生还款失败异及等待还款的提单，重新补还款机制(还款失败，代表回调成功，但还款在写入资料库时异常)
	 * 备注：每3分钟处理一次还款
	 */
	//c.AddJob("0 0/3 * * * * ?", //3分鐘
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.HandleRepayment{}),
	//)

	/**
	代付交易中的单，2分钟没有回调则通知警讯
	*/
	//c.AddJob("0 0/2 * * * * ", //2分鐘
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.NotifyProxyOrder{}),
	//)

	/**
	 * 处里超过5分钟渠道尚未回调的代付提单
	 * (备注：主动前往渠道查询结果，如果已确认为成功或失败，则透过回调机制写回diorpayment)
	 * 备注：5分钟处理一次还款
	 */
	//c.AddJob("0 0/5 * * * * ?", //5分鐘
	//	cron.NewChain().
	//		Then(&cronjob.QueryTransaction{}),
	//)

	//整点开始执行 结算商户报表
	//c.AddJob("*/5 * * * * ?", //每5秒
	c.AddJob("0 0 * * * ?", //
		cron.NewChain(cron.SkipIfStillRunning(logger)).
			Then(&cronjob.MerhchantReport{}),
	)

	// (補算傭金利潤Schedule) 整點開始每5分鐘執行
	//c.AddJob("0 0/5 * * * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.CalculateProfit{}),
	//)
	//
	//// (計算月傭金報表Schedule) 每月2號 03:00:00執行
	//c.AddJob("0 0 3 2 * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.CommissionMonthReport{}),
	//)
	//
	//// (計算月收益報表Schedule) 每月5號 03:00:00執行 '
	//c.AddJob("0 0 3 5 * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.MonthProfitReport{}),
	//)
	//
	//// (查询渠道馀额Schedule) 整點開始每5分鐘執行 '
	//c.AddJob("0 0/5 * * * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.ChannelBalance{}),
	//)
	//
	//// (查询渠道馀额紀錄 Schedule) 整點開始執行 '
	//c.AddJob("0 0 * * * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.ChannelBalanceRecord{}),
	//)
	//
	//// (检查商户子钱包馀额Schedule) 整點開始每10分鐘執行 '
	//c.AddJob("0 0/10 * * * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.MerchantBalancesCheck{}),
	//)

	// (搬移资料到备份表 Schedule) 每日5點開始執行 '
	//c.AddJob("0 0 5 * * ?",
	//	cron.NewChain(cron.SkipIfStillRunning(logger)).
	//		Then(&cronjob.BackupData{}),
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
