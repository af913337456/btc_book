package queue

import (
	"testing"
	"github.com/btc_book/rpc"
	"os"
	"os/signal"
	"fmt"
	"time"
)

func TestBTCMempoolTxParser_IsTxInMempool(t *testing.T) {

	client := rpc.NewBTCRPCHttpClient( // 初始化 rpc 客户端
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")

	parser := NewBTCMempoolTxParser(client) // 初始化解析器
	parser.Start() // 启动解析器

	listenSysInterrupt(func() { // 监听系统中断
		parser.Stop() // 在回调函数中，优雅停止解析器
	})

	targetSearchTxHash := ""
	go func() {
		time.Sleep(time.Second * 12)
		targetSearchTxHash = "df760a7c35f43648320de8a99f5d903c895d793c7b2c8f0c86e6c99e7f3f35cc"
	}()
	go func() {
		for {
			// IsTxInMempool 是查询交易是否在内存池的函数
			if exist := parser.IsTxInMempool(targetSearchTxHash); exist {
				fmt.Println("查询存在:",targetSearchTxHash)
			} else {
				fmt.Println("查询不存在:",targetSearchTxHash)
			}
			time.Sleep(time.Second * 10) // 模拟每隔 10 秒请求一次服务器
		}
	}()

	select {} // 模拟主线程 main 函数的阻塞
}

func listenSysInterrupt(callback func())  {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	go func() {
		for {
			select {
			case <-signalChan:
				fmt.Println("捕获到中断信号")
				callback() // 进行回调
				os.Exit(1)
			}
		}
	}()
}













