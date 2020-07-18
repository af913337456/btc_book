package dao

// 比特币交易，对应于数据库表的交易数据结构体
type Transaction struct {
	Id         int64  `json:"id"` // 主键
	Block_hash string `json:"block_hash"`
	TxId       string `json:"tx_id"`
	VinSize    int64  `json:"vin_size"`
	VoutSize   int64  `json:"vout_size"`
	PackTime   int64  `json:"pack_time"`
	Fork       bool   `json:"fork"`
	Height     int64  `json:"height"`
	TxHash     string `json:"tx_hash"`
}
