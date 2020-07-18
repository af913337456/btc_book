package queue
import (
	"github.com/btc_book/rpc"
	"fmt"
	"time"
	"github.com/btc_book/mempool/syncmap"
)
const (
	MAX_QUEUE_SIZE = 500 // 定义队列的最大元素个数为 500
)

type BTCMempoolTxParser struct {
	RpcClient 	 *rpc.BTCRPCClient // rpc 客户端指针对象
	mempoolTxMap syncmap.Map       // 并发安全的 map，用来存储内存池的交易 hash
	Link         *LinkNode         // 实现队列的链表
	stop         chan bool         // 控制安全退出的无缓冲类型布尔管道变量
}

func NewBTCMempoolTxParser(rpcClient *rpc.BTCRPCClient) *BTCMempoolTxParser {
	queue := BTCMempoolTxParser{
		RpcClient: rpcClient,
		stop:      make(chan bool),  // 实例化控制管道
		mempoolTxMap: syncmap.Map{}, // 实例化 Map
	}
	return &queue // 返回指针对象
}

func (q *BTCMempoolTxParser) IsTxInMempool(txHashStr string) bool {
	val,_ := q.mempoolTxMap.Load(txHashStr) // 直接从 map 中加载判断
	return val != nil
}

// 在子协程中，进行交易内存池的交易数据分析，定时去读取节点内存池数据
func (q *BTCMempoolTxParser) Start() {
	go func() {
		for { // 死循环
			select {
			case <- q.stop: // 如果监听到退出监控，那么终止队列的数据循环分析
				fmt.Println("stop event happened, exit BTCMempoolTxParser....")
				return
			default:
				txHashArray,err := q.RpcClient.GetRpc().GetRawMempool() // 调用 RPC 接口
				if err != nil {
					fmt.Println("GetRawMempoolVerbose err:",err.Error())
					time.Sleep(time.Second * 5) // 如果接口返回错误，延迟 5 秒继续
					continue
				}
				fmt.Println("获取一次内存池交易，結果是:",txHashArray)
				// txHashArray 是内存池交易的 txHash 数组
				for _,txHash := range txHashArray {
					txHashStr := txHash.String()
					// Load 进行存储
					if val,_ := q.mempoolTxMap.Load(txHashStr); val != nil {
						// 这条记录已经存在，那么跳过它，开始下一轮
						continue
					}
					// 接着判断是否超过最大队列数
					if q.mempoolTxMap.Size() >= MAX_QUEUE_SIZE {
						// RemoveFirst 队伍满了，要出队，先被插入队伍的，是最早的记录，从队伍头出队
						removeHash := q.Link.RemoveFirst()
						// mempoolTxMap.Delete 同步到 map 结构中，把已经出队的取出
						q.mempoolTxMap.Delete(removeHash)
					}
					q.Link.AddAtTail(txHashStr) // 新纪录从队尾入队
					fmt.Println("入队一条:",txHashStr)
					q.mempoolTxMap.Store(txHashStr,byte(1)) // 同步存储到 map 中，方便查询
				}
				time.Sleep(time.Second * 10) // 设置为 10 秒一次查询
			}
		}
	}()
}

// stop 停止队列分析交易内存池
func (q *BTCMempoolTxParser) Stop() {
	q.stop <- true
}























