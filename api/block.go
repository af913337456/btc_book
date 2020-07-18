package api

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// 根据区块哈希获取区块信息
func GetBlockInfoByBlockHash(client *rpcclient.Client,blockHash string) (*wire.MsgBlock,error) {
	hash,err := chainhash.NewHashFromStr(blockHash) // 如果不是 16 进制的哈希值，这里转换会报错
	if err != nil {
		return nil,err
	}
	return client.GetBlock(hash)
}

// 获取链的最新区块高度
func GetBlockCount(client *rpcclient.Client) (int64,error) {
	return client.GetBlockCount()
}

// 根据区块高度区块哈希
func GetBlockHashByBlockHeight(client *rpcclient.Client,height int64) (*chainhash.Hash,error) {
	return client.GetBlockHash(height)
}
















