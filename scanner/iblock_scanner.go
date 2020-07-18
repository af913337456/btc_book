package scanner

import (
	"math/big"
	"github.com/go-xorm/xorm"
)

// 描述：作为兼容多条链的区块扫描，独立出公共的数据结构 与 接口

// ScannerBlockInfo 是区块扫描时候，会用到的公共数据结构体
type ScannerBlockInfo struct {
	BlockHash   string `json:"block_hash"`
	ParentHash  string `json:"parent_hash"`
	BlockNumber string `json:"block_number"`
	Timestamp   string `json:"timestamp"`
	Txs         interface{} `json:"txs"`
}

// IBlockScanner 是一个 interface 类型，不是 struct，它是 Go 语言中的接口
type IBlockScanner interface {

	// 获取父区块 hash 值的接口函数，根据 childHash 子哈希为参数，第一个返回值是父哈希
	GetParentHash(childHash string) (string,error)

	// 获取最新的区块的号
	GetLatestBlockNumber() (*big.Int, error)

	// 根据区块号获取区块的信息，返回的信息是公共数据结构体
	GetBlockInfoByNumber(blockNumber *big.Int) (*ScannerBlockInfo, error)

	// 插入区块内的交易数据到数据库，因为交易 transaction 结构体，不同的链，存在差异性，所以这里定义的是 interface 泛型
	InsertTxsToDB(blockNumber int64,tx *xorm.Session,transactions interface{}) error

	// 交易数据处理函数，和 InsertTxsToDB 不同，我们可以在交易数据入库后，再做特定的解析处理，比如过滤数据
	TransactionHandler(txs interface{})
}















