package rpc

import (
	"testing"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"encoding/json"
)

func Test_RPCClient(t *testing.T) {
	client := NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	bestBlockHash,err := client.GetRpc().GetBestBlockHash()
	if err != nil {
		t.Log(err.Error())
	}else {
		t.Log(bestBlockHash.String())
	}
}

func Test_RPCClient_Count(t *testing.T) {
	client := NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	blockNum,err := client.GetRpc().GetBlockCount()
	if err != nil {
		t.Log(err.Error())
	}else {
		t.Log(blockNum)
	}
}

func Test_RPCClient_Block(t *testing.T) {
	client := NewBTCRPCHttpClient(
		"127.0.0.1:8332",
		"mybtc",
		"mypassword")
	hash,_ := chainhash.NewHashFromStr("44280d612f7ee1ce8c9d6e2ee5b47547396863d91d8580279dfca43bb6c0bf4e")
	block,err := client.GetRpc().GetBlockVerbose(hash)
	if err != nil {
		t.Log(err.Error())
	}else {
		bys,_ := json.Marshal(block)
		t.Log(string(bys))
	}
}

