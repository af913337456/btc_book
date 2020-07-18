package scanner

import (
	"math/big"
	"github.com/go-xorm/xorm"
	"github.com/btc_book/rpc"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"errors"
	"fmt"
	"strconv"
	"github.com/btc_book/dao"
	"time"
	"encoding/hex"
	"encoding/base64"
	"strings"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2020/6/7
*/


type LocalBitcoindScanner struct {
	// 因为要访问节点的RPC接口，这里需要定义个RPC客户端
	rpcClient *rpc.BTCRPCClient
}

// 以比特币 rpc 客户度初始化
func NewLocalBitcoindScanner(rpcClient *rpc.BTCRPCClient) *LocalBitcoindScanner {
	return &LocalBitcoindScanner{rpcClient: rpcClient}
}

func (client *LocalBitcoindScanner) GetParentHash(childHash string) (string,error) {
	hash,err := chainhash.NewHashFromStr(childHash)
	if err != nil {
		return "",errors.New("invalid childHash")
	}
	// GetBlockHeaderVerbose 根据区块的哈希获取区块的头部数据
	blockHeader,err := client.rpcClient.GetRpc().GetBlockHeaderVerbose(hash)
	if err != nil {
		return "",fmt.Errorf("GetParentHash GetBlockHeaderVerbose err: %s",err.Error())
	}
	// PreviousHash 就是前一个区块的 hash，即父 hash
	return blockHeader.PreviousHash,nil
}

// 考虑到区块链中涉及到的大数比较多，所以在返回区块号这里，
// 我们统一使用 Go 汇总的 big.Int，大整数作为返回类型
func (client *LocalBitcoindScanner) GetLatestBlockNumber() (*big.Int, error){
	blockNumber,err := client.rpcClient.GetRpc().GetBlockCount()
	if err != nil {
		return nil,fmt.Errorf("GetLatestBlockNumber GetBlockCount err: %s",err.Error())
	}
	return big.NewInt(blockNumber),nil
}

// 根据区块号，获取区块信息
func (client *LocalBitcoindScanner) GetBlockInfoByNumber(blockNumber *big.Int) (*ScannerBlockInfo, error){
	if blockNumber == nil {
		return nil,errors.New("invalid blockNumber")
	}
	// 这里要分3步完成：
	// 1. 先根据区块号，获取区块的 hash 值；
	// 2. 再根据 hash 值获取区块的信息;
	// 3. 最后根据区块内的交易hash 去获取交易信息。
	rpcClient := client.rpcClient.GetRpc()
	blockHash,err := rpcClient.GetBlockHash(blockNumber.Int64())
	if err != nil {
		return nil,fmt.Errorf("GetBlockInfoByNumber GetBlockHash err: %s",err.Error())
	}
	block,err := rpcClient.GetBlockVerbose(blockHash)
	if err != nil {
		return nil,fmt.Errorf("GetBlockInfoByNumber GetBlockVerbose err: %s",err.Error())
	}
	// block.RawTx 其实返回的是空数组，我们不要使用它，用 block.Tx
	// 下面的操作比较耗时
	txList := []dao.Transaction{}
	for _, tx := range block.Tx {
		txHash,_ := chainhash.NewHashFromStr(tx)
		txObject,err := rpcClient.GetRawTransactionVerbose(txHash)
		if err != nil {
			// 打印错误并继续
			fmt.Println(fmt.Errorf("GetBlockInfoByNumber GetRawTransactionVerbose err: %s",err.Error()).Error())
			continue
		}
		item := dao.Transaction{
			TxId:       txObject.Txid,
			Block_hash: block.Hash,
			VinSize:    int64(len(txObject.Vin)),
			VoutSize:   int64(len(txObject.Vout)),
			PackTime:   txObject.Time,
			Fork:       false, // 开始的，默认都不是分叉块，所以是 false
			Height:     block.Height,
			TxHash:     txObject.Hash,
		}
		txList = append(txList,item)
	}
	return &ScannerBlockInfo{
		BlockHash:   block.Hash,
		ParentHash:  block.PreviousHash,
		BlockNumber: strconv.FormatInt(block.Height,10), // int64 转 string 字符串
		Timestamp:   strconv.FormatInt(block.Time,10),
		Txs:         txList, // 将交易数据数组赋值进去
	},nil
}

// 将交易数据插入到数据库
func (client *LocalBitcoindScanner) InsertTxsToDB(blockNumber int64,tx *xorm.Session,transactions interface{}) error {
	// 下面使用 Go 的语法，将 interface 强转为 []dao.Transaction
	txs := transactions.([]dao.Transaction) // 对应于 GetBlockInfoByNumber 函数最后返回的 ScannerBlockInfo 中的 Txs
	if _, err := tx.Insert(&txs); err != nil { // 插入
		tx.Rollback()  // 事务回滚
		return err
	}
	return nil
}

// 处理交易数据
func (client *LocalBitcoindScanner) TransactionHandler(txs interface{}){
	// 这里实现对应的交易解析，等待实现
	realTxs := txs.([]dao.Transaction)
	rpcClient := client.rpcClient.GetRpc()
	go func() { // 遍历交易，是个耗时的过程，这里我们考虑启动协程异步处理
		fmt.Println("要处理的交易数: ",len(realTxs))
		for _, tx := range realTxs {
			// 先根据交易的 id 获取交易的具体数据
			txHash,_ := chainhash.NewHashFromStr(tx.TxId)
			retryCounter := 0 // 控制重试次数
			RETRY:
			txData,err := rpcClient.GetRawTransactionVerbose(txHash)
			if err != nil {
				// 记录错误，并延迟一会后重试
				fmt.Println("TransactionHandler GetRawTransactionVerbose err: ",tx.TxId,err.Error())
				time.Sleep(time.Millisecond * 500)
				retryCounter ++
				if retryCounter <= 3 {
					goto RETRY
				}
				continue // 如果错误次数太多，那么就跳过这一条
			}
			fmt.Println(fmt.Sprintf("被分析的交易: %s，输出条数: %d",tx.TxId,len(txData.Vout)))
			// 开始遍历 vout
			for index, out := range txData.Vout {
				if out.ScriptPubKey.Type != "nulldata" {
					fmt.Println(fmt.Sprintf("非 OP_RETURN 类型，下标：%d",index))
					continue // 跳过不是的
				}
				// 如果脚本的类型是空字符串，那么这是一个 OP_RETURN 类型的输出，符合条件
				opreturnData := strings.Split(out.ScriptPubKey.Asm," ")[1] // 这里取出交易的数据: opreturn <data>
				dataBytes,err := hex.DecodeString(opreturnData) // 第一次解码，是比特币默认的方式
				if err != nil {
					fmt.Println("TransactionHandler GetRawTransactionVerbose 非法交易: ",err.Error())
					continue
				}
				// 然后根据自定义的解码方式，进行数据解码，开始恢复原始的数据
				// 下面假设我们的编码数据方式是 base64 的方式
				encData := string(dataBytes)
				decBytes,err := base64.StdEncoding.DecodeString(encData)
				if err != nil {
					// 如果解码失败，那么就证明这不是由我们系统发出的交易，自然就是跳过
					fmt.Println("TransactionHandler GetRawTransactionVerbose 不是由系统发出的交易: ",err.Error())
					continue
				}
				originData := string(decBytes)
				// 对数据进行一些其他的处理，比如记录等操作，这里我们直接打印出即可
				fmt.Println("解析出的 OP_RETURN 数据是: ",originData)
			}
		}
		fmt.Println("处理结束")
	}()
}
















