package api

import (
	"testing"
	"github.com/btc_book/rpc" // 要注意，这里的 rpc 包，要使用我们当前项目提供的
	"encoding/json"
)

func Test_GetBlockInfoByBlockHash(t *testing.T) {
	client := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	bestBlockHash,err := client.GetRpc().GetBestBlockHash()
	if err != nil {
		t.Log(err.Error())
	}else {
		t.Log(bestBlockHash.String())
		// 获取了区块的哈希数据后，进行区块数据的获取
		blockInfo,err := GetBlockInfoByBlockHash(client.GetRpc(),bestBlockHash.String())
		if err != nil {
			t.Log("GetBlockInfoByBlockHash err:",err.Error())
		}else {
			bytes,_ := json.Marshal(blockInfo) // 将结构体数据转 json
			t.Log(string(bytes))
		}
	}
}

func Test_GetBlockCount(t *testing.T) {
	client := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	blockHeight,err := GetBlockCount(client.GetRpc())
	if err != nil {
		t.Log("GetBlockCount err:",err.Error())
	}else {
		t.Log("blockHeight ===> ",blockHeight)
	}
}

func Test_GetBlockHashByBlockHeight(t *testing.T) {
	client := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	// 获取高度为 1 的区块哈希
	blockHash,err := GetBlockHashByBlockHeight(client.GetRpc(),1)
	if err != nil {
		t.Log("GetBlockCount err:",err.Error())
	}else {
		// 使用 String() 方法输出结果
		t.Log("blockHash ===> ",blockHash.String())
	}
}

func Test_GetTransactionInfoByTxHash(t *testing.T) {
	client := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	bestBlockHash,err := client.GetRpc().GetBestBlockHash()
	if err != nil {
		t.Log(err.Error())
	}else {
		t.Log(bestBlockHash.String())
		// 获取了区块的哈希数据后，进行区块数据的获取
		blockInfo,err := GetBlockInfoByBlockHash(client.GetRpc(),bestBlockHash.String())
		if err != nil {
			t.Log("GetBlockInfoByBlockHash err:",err.Error())
		}else {
			bytes,_ := json.Marshal(blockInfo) // 将结构体数据转 json
			t.Log(string(bytes))
			// 开始获取交易数据
			if len(blockInfo.Transactions) > 0 { // 如果该区块有交易数据，才进入 if 内部
				// 提取第一条交易用作测试
				firstTxHashStr := blockInfo.Transactions[0].TxHash().String()
				txData,err := GetTransactionInfoByTxHash(client.GetRpc(),firstTxHashStr)
				if err != nil {
					t.Log("GetBlockInfoByBlockHash err:",err.Error())
				}else{
					bytes,_ := json.Marshal(txData)
					t.Log("tx data:",string(bytes))
				}
			}
		}
	}
}









