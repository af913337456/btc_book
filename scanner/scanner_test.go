package scanner

import (
	"testing"
	"github.com/btc_book/dao"
	"github.com/btc_book/rpc"
	"runtime"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2020/6/7
*/

// 测试比特币区块遍历
func TestBlockScanner_Start(t *testing.T) {
	runtime.GOMAXPROCS(4)
	// 初始化 btc RPC 客户端
	rpcClient := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")

	// 初始化 btc 遍历者
	client := NewLocalBitcoindScanner(rpcClient)

	// 初始化数据库连接者的配置对象。记得修改为你本地数据库的参数
	option := dao.MysqlOptions{
		Hostname:           "127.0.0.1",
		Port:               "3306",
		DbName:             "btc_relay",
		User:               "root",
		Password:           "123aaa",
		TablePrefix:        "btc_",
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
	}
	// 添加表格
	tables := []interface{}{}
	tables = append(tables, dao.Block{}, dao.Transaction{})
	// 根据上面定义的配置，初始化数据库连接者
	mysql := dao.NewMqSQLConnector(&option, tables)
	// 初始化区块扫描者
	scanner := NewBlockScanner(client, mysql)
	// 设置固定的开始扫描高度
	// scanner.SetStartScannerHeight(56149) // 这里会报错，因为我的本地节点区块高度还没这么高
	err := scanner.Start() // 开始扫描
	if err != nil {
		panic(err)
	}
	// 使用 select 模拟阻塞主协程，等待上面的代码执行，因为我们的扫描是在 gorutine 协程中进行的
	select {}
}




















