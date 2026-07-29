package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(logx.LogConf{
		ServiceName: "async_account_check",
		Mode:        "console",
	})

	db := sqlx.MustNewConn(c.DB)

	accountMgrClient := NewAccountMgrClient(c.AccountMgrRpc)
	model := NewC2cPendingTransferModel(db)
	checker := NewC2CChecker(accountMgrClient, model)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(c.CheckInterval) * time.Second)
	defer ticker.Stop()

	fmt.Printf("Starting c2c async account checker...\n")
	fmt.Printf("Check interval: %d seconds\n", c.CheckInterval)
	fmt.Printf("Time range: -%d seconds to -%d seconds\n", -c.TimeStart, -c.TimeEnd)

	go func() {
		for {
			select {
			case <-ticker.C:
				runCheck(checker, c.TimeStart, c.TimeEnd)
			case <-ctx.Done():
				return
			}
		}
	}()

	//先检查一次，之后再根据配置间隔检查
	runCheck(checker, c.TimeStart, c.TimeEnd)

	// 只要收到信号SIGINT或SIGTERM，就退出
	<-sigChan
	fmt.Println("Shutting down...")
}

func runCheck(checker *C2CChecker, timeStart, timeEnd int64) {
	batchStart := time.Now()
	startTime := batchStart.Add(time.Duration(timeStart) * time.Second)
	endTime := batchStart.Add(time.Duration(timeEnd) * time.Second)

	fmt.Printf("\n[%s] Starting check batch...\n", batchStart.Format("2006-01-02 15:04:05"))

	result := checker.Check(context.Background(), startTime, endTime)
	result.ExecuteStartTime = batchStart
	result.ExecuteEndTime = time.Now()

	fmt.Println(result.String())
}
