package cronjob

import (
	"fmt"
	"github.com/copo888/copo_schedule/helper"
	"github.com/neccoys/go-zero-extension/redislock"
	"time"
)

type TestJob struct {
}

func (a *TestJob) Run() {
	// handle

	// redis test
	//err := helper.REDIS.Set(context.Background(), "testtest", "111111", 0).Err()
	//if err != nil {
	//	panic(err)
	//}
	//
	//v, err := helper.REDIS.Get(context.Background(), "testtest").Result()
	//
	//if err != nil {
	//	log.Panicln("==========")
	//}
	//
	//fmt.Println(">>>>>>", v)

	for i := 0; i < 3; i++ {
		go func() {
			redisLock := redislock.New(helper.REDIS, "testtest", "test_")
			redisLock.SetExpire(3)

			if ok, _ := redisLock.TryLockTimeout(3); ok {
				defer redisLock.Release()

				fmt.Println("cccc")

				list := []struct {
					MerchantCode string
					PaytypeCode  string
					Fee          float64
					HandlingFee  float64
				}{}

				helper.COPO_DB.Table("mc_merchant_channel_rate").
					Where("merchant_code = ?", "ME00001").
					Find(&list)

				fmt.Println(list)
				time.Sleep(3 * time.Second)

				//log.Println(">>>>>>>>>", "lock.....")
			}
		}()
	}

	//
	time.Sleep(10 * time.Second)
}
