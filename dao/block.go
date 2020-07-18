package dao

// 存储区块信息的区块结构体，如果需要定义索引，可以查看xorm库的标注语法，进行拓展
type Block struct {
	Id          int64  `json:"id"`           // 主键
	BlockNumber string `json:"block_number"` // 区块号
	BlockHash   string `json:"block_hash"`   // 区块 hash
	ParentHash  string `json:"parent_hash"`  // 夫区块 hash
	CreateTime  int64  `json:"create_time"`  // 区块的生成时间
	Fork        bool   `json:"fork"`         // 是否是分叉区块
}
