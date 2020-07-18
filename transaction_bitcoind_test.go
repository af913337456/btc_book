package btc_book

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2020/6/20
*/


import (
	"testing"
	"github.com/btc_book/rpc"
	"fmt"
	"encoding/base64"
)

func TestSendLocalNet_BTCNormalTransaction(t *testing.T) {
	rpcClient := rpc.NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	// 这里我们发送 3 笔交易
	for  i:=0; i< 1; i++ {
		// 下面的数据编码方式，要对应到 TransactionHandler 函数中的 解码 方式
		dataBytes := []byte(fmt.Sprintf("你好，这是我交易的备注信息，下标: %d", i))
		encData := base64.StdEncoding.EncodeToString(dataBytes)
		fmt.Println("加密的数据是：",encData)
		err := SendLocalNode_BTCNormalTransaction(
			rpcClient,
			"cRgsH3pQMVhdux6HGzyMeRkoESfqtUNe5GhEvNw8Er7f4jTJwuoL", // 发送者私钥
			"mwzzVeqDfFXFD1vn6jsTJyakeaSA6kRjen", // 接收交易的地址
			200,&OpReturnDataObj{
				Data: encData, // op_return 携带编码后的数据
			})
		t.Log(err)
	}
}