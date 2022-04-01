package cronjob

import (
	"fmt"
	"time"
)

type TestJob struct {
}

func (a *TestJob) Run() {
	// handle
	fmt.Println("cccc")
	time.Sleep(10 * time.Second)
}
