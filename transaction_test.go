package btc_book

import (
	"testing"
	"encoding/hex"
	"encoding/base64"
	"fmt"
)

func TestGetUTXO(t *testing.T) {
	ret,err := getUTXOListFromBlockCypherAPI(
		"mq5kKXGw8ERWJQ2HLLunkh3FTBk281ZC47",
		"btc/test3")
	t.Log(err) // 打印是否与错误信息
	t.Log(ret) // 打印出结果
}


func TestSendTestNet_BTCNormalTransaction(t *testing.T) {
	err := SendTestNet_BTCNormalTransaction(
		"KzZkYh62v6xq2SdMaYbuR6yhbbav1Pq9cXGU6M8Ci8m6J6qc23r3",
		"1KFHE7w8BhaENAswwryaoccDb6qcT6DbYY",
		2000,nil)
	t.Log(err)
}

func TestSendTestNet_BTCNormalTransaction_withOpReturnOutput(t *testing.T) {
	err := SendTestNet_BTCNormalTransaction(
		"KzZkYh62v6xq2SdMaYbuR6yhbbav1Pq9cXGU6M8Ci8m6J6qc23r3",
		"1KFHE7w8BhaENAswwryaoccDb6qcT6DbYY",
		200,&OpReturnDataObj{
			Data: "你好，这是我交易的备注信息",
		})
	t.Log(err)
}

func TestSendTestNet_BTCNormalTransaction_withOpReturnOutput_N(t *testing.T) {
	// 这里我们发送 3 笔交易
	for  i:=0; i< 3; i++ {
		// 下面的数据编码方式，要对应到 TransactionHandler 函数中的 解码 方式
		dataBytes := fmt.Sprintf("你好，这是我交易的备注信息，下标: %d", i)
		encData := base64.StdEncoding.EncodeToString([]byte(dataBytes))
		err := SendTestNet_BTCNormalTransaction(
			"KzZkYh62v6xq2SdMaYbuR6yhbbav1Pq9cXGU6M8Ci8m6J6qc23r3",
			"1KFHE7w8BhaENAswwryaoccDb6qcT6DbYY",
			200,&OpReturnDataObj{
				Data: encData, // op_return 携带编码后的数据
			})
		t.Log(err)
	}
}

func Test_DecodeOpreturnData(t *testing.T) {
	hexData := "e4bda0e5a5bdefbc8ce8bf99e698afe68891e4baa4e69893e79a84e5a487e6b3a8e4bfa1e681af"
	originDataBytes,_ := hex.DecodeString(hexData)
	t.Log(string(originDataBytes))
}









