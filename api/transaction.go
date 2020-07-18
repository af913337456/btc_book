package api

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/btcjson"
)

// 根据交易哈希获取交易数据
func GetTransactionInfoByTxHash(client *rpcclient.Client,txHash string) (*btcjson.TxRawResult,error) {
	hash,err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil,err
	}
	return client.GetRawTransactionVerbose(hash)
}









