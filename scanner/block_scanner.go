package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
	"github.com/btc_book/dao"
)

/**
  author : LinGuanHong
  github : https://github.com/af913337456
  blog   : http://www.cnblogs.com/linguanh
  time   : 14:49
*/

// DES: 区块扫描者。遍历出区块的交易，方便从交易中解析出数据，做自定义操作
// 区块遍历者


type BlockScanner struct {
	client       IBlockScanner      // 接口实现者
	mysql        dao.MySQLConnector // 数据库连接者对象
	lastBlock    *dao.Block         // 用来存储每次遍历后，上一次的区块
	lastNumber   *big.Int           // 上一次区块的区块号
	fork         bool               // 区块分叉标记位
	stop         chan bool          // 用来控制是否停止遍历的管道
	lock         sync.Mutex         // 互斥锁，控制并发
	delay        time.Duration      // 扫描的间隔时间
}

func NewBlockScanner(scanner IBlockScanner, mysql dao.MySQLConnector) *BlockScanner {
	return &BlockScanner{
		client:       scanner,
		mysql:        mysql,
		lastBlock:    &dao.Block{},
		fork:         false,
		stop:         make(chan bool),
		lock:         sync.Mutex{},
		delay:        time.Duration(30 * time.Second),
	}
}

var compare = func(arg1,arg2 string) bool {
	if arg1 == "" {
		return false
	}
	if arg2 == "" {
		return true
	}
	a,_ := new(big.Int).SetString(arg1, 10)
	b,_ := new(big.Int).SetString(arg2, 10)
	return a.Cmp(b) >= 0
}

// 初始化：内部在开始遍历时赋值 lastBlock
func (scanner *BlockScanner) init() error {
	// 下面使用 xorm 提供的数据库函数来
	// 从数据库中寻找出上一次成功遍历的且不是分叉的区块
	// 等同于SQL：select * from eth_block where fork=false order by create_time desc limit 1;
	setNextBlockNumber := func() {
		// 区块 hash 不为空，证明不是首次启动了。是后续的启动
		scanner.lastNumber, _ = new(big.Int).SetString(scanner.lastBlock.BlockNumber, 10)
		// 下面加 1，因为上一次数据库存的是已经遍历完了的，接下来的是它的下一个
		scanner.lastNumber.Add(scanner.lastNumber, new(big.Int).SetInt64(1))
	}
	getDbLastBlock := func() (*dao.Block,error) {
		dbBlock := &dao.Block{}
		_, err := scanner.mysql.Db.
			Desc("create_time"). // 根据时间倒叙
			Where("fork = ?", false).
			Get(dbBlock)
		if err != nil {
			return nil,err
		}
		return dbBlock,nil
	}
	dbBlock,err := getDbLastBlock()
	if err != nil {
		return err
	}
	if scanner.lastBlock.BlockHash != "" {
		// 被设置了的情况
		// 与 db 的比较，找出最新的
		fmt.Println(dbBlock.BlockNumber)
		if compare(dbBlock.BlockNumber,scanner.lastBlock.BlockNumber) {
			// 使用 db 的
			scanner.lastBlock = dbBlock
		}
		setNextBlockNumber()
		return nil
	}
	scanner.lastBlock = dbBlock
	if scanner.lastBlock.BlockHash == "" {
		// 区块 hash 为空，证明是整个程序的首次启动，那么从节点中获取最新生成的区块
		// GetLatestBlockNumber 获取最新区块的区块号
		latestBlockNumber, err := scanner.client.GetLatestBlockNumber()
		if err != nil {
			return err
		}
		// GetBlockInfoByNumber 根据区块号获取区块数据
		latestBlock, err := scanner.client.GetBlockInfoByNumber(latestBlockNumber)
		if err != nil {
			return err
		}
		if latestBlock.BlockNumber == "" {
			panic(latestBlockNumber.String())
		}
		// 下面是赋值区块遍历者的 lastBlock 变量
		scanner.lastBlock.BlockHash   = latestBlock.BlockHash
		scanner.lastBlock.ParentHash  = latestBlock.ParentHash
		scanner.lastBlock.BlockNumber = latestBlock.BlockNumber
		scanner.lastBlock.CreateTime  = scanner.hexToTen(latestBlock.Timestamp).Int64()
		scanner.lastNumber = latestBlockNumber
	} else {
		setNextBlockNumber()
	}
	return nil
}

// 设置固定的开始高度，起效要求要在开始 scan 之前
func (scanner *BlockScanner) SetStartScannerHeight(height int64) {
	blockInfo,err := scanner.retryGetBlockInfoByNumber(new(big.Int).SetInt64(height))
	if err != nil {
		panic(fmt.Errorf("指定区块高度出错，请检查数值是否超过当前链节点的区块高度，rawErrInfo %s",err.Error()))
	}
	scanner.lastBlock.BlockHash   = blockInfo.BlockHash
	scanner.lastBlock.ParentHash  = blockInfo.ParentHash
	scanner.lastBlock.BlockNumber = blockInfo.BlockNumber
	scanner.lastBlock.CreateTime  = scanner.hexToTen(blockInfo.Timestamp).Int64()
}

// 整个区块扫描的启动函数
func (scanner *BlockScanner) Start() error {
	scanner.lock.Lock()  // 互斥锁加锁，在 stop 函数内有解锁步骤
	// 首先调用 init 进行数据初始化，内部主要是初始化区块号
	if err := scanner.init(); err != nil {
		scanner.lock.Unlock() // 因为出现了错误，我们要进行解锁
		return err
	}
	execute := func() {
		// scan 函数，就是区块扫描函数
		if err := scanner.scan(); nil != err {
			scanner.log("scanner err :",err.Error())
			return
		}
		time.Sleep(1 * time.Second) // 延迟一秒开始下一轮
	}
	// 启动一个 go协程 来遍历区块
	go func() {
		for {
			select {
			case <-scanner.stop: // 监听是否退出遍历
				scanner.log("finish block scanner!")
				return
			default:
				if !scanner.fork {
					// 进入这个 if 证明没有检测到分叉。正常地进行每一轮的遍历
					execute()
					continue
				}
				// fork = true，则监听到有分叉，重新初始化
				// 重新从数据库获取上次遍历成功的且没有分叉的区块号
				if err := scanner.init(); err != nil {
					scanner.log(err.Error())
					return
				}
				scanner.fork = false
			}
		}
	}()
	return nil
}

// 公有函数，可以共外部调用，来控制停止区块遍历
func (scanner *BlockScanner) Stop() {
	scanner.lock.Unlock()  // 解锁
	scanner.stop <- true
}

// 输出日志
func (scanner *BlockScanner) log(args ...interface{}) {
	fmt.Println(args...)
}

// 是否分叉，返回 true 是分叉
func (scanner *BlockScanner) isFork(currentBlock *dao.Block) bool {
	if currentBlock.BlockNumber == "" {
		panic("invalid block")
	}
	// scanner.lastBlock.BlockHash == currentBlock.ParentHash 上一次的区块 hash 是否是 当前区块的父区块 hash
	if scanner.lastBlock.BlockHash == currentBlock.BlockHash || scanner.lastBlock.BlockHash == currentBlock.ParentHash {
		scanner.lastBlock = currentBlock // 没有发生分叉，则更新上一次区块为当前被检测的
		return false
	}
	return true
}

func (scanner *BlockScanner) forkCheck(currentBlock *dao.Block) bool {
	if !scanner.isFork(currentBlock) {
		return false
	}
	// 获取出最初开始分叉的那个区块
	forkBlock, err := scanner.getStartForkBlock(currentBlock.ParentHash)
	if err != nil {
		panic(err)
	}
	scanner.lastBlock = forkBlock // 更新。从这个区块开始分叉的
	numberEnd := ""
	if strings.HasPrefix(currentBlock.BlockNumber, "0x") {
		// 16 进制转为 10 进制
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber[2:], 16)
		numberEnd = c.String()
	} else {
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber, 10)
		numberEnd = c.String()
	}
	numberFrom := forkBlock.BlockNumber
	// 下面使用 xorm 提供的函数执行数据库更新操作：
	// 将范围内的区块将分叉标志位设置位分叉
	_, err = scanner.mysql.Db.
		Table(dao.Block{}).
		Where("block_number > ? and block_number < ?", numberFrom, numberEnd).
		Update(map[string]bool{"fork": true})
	if err != nil {
		panic(fmt.Errorf("update fork block failed %s", err.Error()))
	}
	return true
}

// 获取分叉点区块
func (scanner *BlockScanner) getStartForkBlock(parentHash string) (*dao.Block, error) {
	// 获取当前区块的父区块，分叉从父区块开始
	parent := dao.Block{} // 定义一个 block 结构体实例，用来存储从数据库查询出的区块信息
	// 下面使用 xorm 框架提供的函数，根据 block_hash 去数据库获取区块信息，等同于 SQL 语句：
	// select * from eth_block where block_hash=parentHash limit 1;
	_, err := scanner.mysql.Db.Where("block_hash=?", parentHash).Get(&parent)
	if err == nil && parent.BlockNumber != "" {
		return &parent, nil  // 本地存在，直接返回分叉点区块
	}
	// 数据库没有父区块记录，准备从以太坊接口获取
	fatherHash, err := scanner.retryGetBlockInfoByHash(parentHash)
	if err != nil {
		return nil, fmt.Errorf("分叉严重错误，需要重启区块扫描 %s", err.Error())
	}
	// 继续递归往上查询，直到在数据库中有它的记录
	return scanner.getStartForkBlock(fatherHash)
}

// 定义一个将16进制转为10进制大数的函数
func (scanner *BlockScanner) hexToTen(hex string) *big.Int {
	if !strings.HasPrefix(hex,"0x") {
		ten, _ := new(big.Int).SetString(hex, 10) // 本身就是 10 进制字符串，直接设置
		return ten
	}
	ten, _ := new(big.Int).SetString(hex[2:], 16)
	return ten
}

// 区块号存在，信息获取为空，可能是以太坊网络延时问题，重试策略函数
func (scanner *BlockScanner) retryGetBlockInfoByNumber(targetNumber *big.Int) (*ScannerBlockInfo, error) {
	Retry:
		// 下面调用我们请求者 client 的 GetBlockInfoByNumber 函数
		fullBlock, err := scanner.client.GetBlockInfoByNumber(targetNumber)
		if err != nil {
			errInfo := err.Error()
			if strings.Contains(errInfo, "empty") || strings.Contains(errInfo, "must retry") {
				// 区块号存在，信息获取为空，可能是以太坊网络延时问题，直接重试
				scanner.log("获取区块信息，重试一次.....", targetNumber.String(),errInfo)
				goto Retry
			}
			return nil, err
		}
	return fullBlock, nil
}

// 区块 hash 存在，信息获取为空，可能是以太坊网络或节点问题，重试策略函数
func (scanner *BlockScanner) retryGetBlockInfoByHash(hash string) (string, error) {
	Retry:
		// 下面调用我们请求者 client 的 GetBlockInfoByHash 函数
		parentHash, err := scanner.client.GetParentHash(hash)
		if err != nil {
			errInfo := err.Error()
			if strings.Contains(errInfo, "empty") || strings.Contains(errInfo, "must retry") {
				// 区块号存在，信息获取为空，可能是以太坊网络延时问题，直接重试
				scanner.log("获取区块信息，重试一次.....", hash,errInfo)
				goto Retry
			}
			return "", err
		}
	return parentHash, nil
}

// 获取要扫描的区块号
func (scanner *BlockScanner) getScannerBlockNumber() (*big.Int,error) {
	// 调用请求者 client 获取公链上最新生成的区块的区块号
	newBlockNumber, err := scanner.client.GetLatestBlockNumber()
	if err != nil {
		return nil,fmt.Errorf("GetLatestBlockNumber: %s",err.Error())
	}
	latestNumber := newBlockNumber
	// 下面使用 new 的形式初始化并设置值，不要直接赋值，
	// 否则会和 lastNumber 的内存地址一样，影响后面的获取区块信息
	targetNumber := new(big.Int).Set(scanner.lastNumber)
	// 比较区块号大小
	// -1 if x <  y，0 if x == y，+1 if x >  y
	if latestNumber.Cmp(scanner.lastNumber) < 0 {
		// 最新的区块高度比设置的要小，则等待新区块高度 >= 设置的
		Next:
			for {
				select {
				case <-time.After(scanner.delay): // 延时4秒重新获取
					number, err := scanner.client.GetLatestBlockNumber()
					if err == nil && number.Cmp(scanner.lastNumber) >= 0 {
						break Next // 跳出循环
					}
				}
			}
	}
	return targetNumber,nil // 返回目标区块高度
}

// 扫描区块
func (scanner *BlockScanner) scan() error {
	// 获取要进行扫描的区块号
	targetNumber,err := scanner.getScannerBlockNumber()
	if err != nil {
		return fmt.Errorf("getScannerBlockNumber: %s",err.Error())
	}
	// 使用具有重试策略的函数获取区块信息
	info, err := scanner.retryGetBlockInfoByNumber(targetNumber)
	if err != nil {
		return fmt.Errorf("retryGetBlockInfoByNumber: %s",err.Error())
	}
	// 区块号自增 1，在下次扫描的时候，指向下一个高度的区块
	scanner.lastNumber.Add(scanner.lastNumber, new(big.Int).SetInt64(1))
	// 因为涉及到两张表的更新，我们需要采用数据库事务处理
	tx := scanner.mysql.Db.NewSession() // 开启事务
	defer tx.Close()
	// 准备保存区块信息，先判断当前区块记录是否已经存在
	block := dao.Block{}
	_, err = tx.Where("block_hash=?", info.BlockHash).Get(&block)
	if err == nil && block.Id == 0 {
		// 不存在，进行添加
		block.BlockNumber = scanner.hexToTen(info.BlockNumber).String()
		block.ParentHash  = info.ParentHash
		block.CreateTime  = scanner.hexToTen(info.Timestamp).Int64()
		block.BlockHash   = info.BlockHash
		block.Fork = false
		if _, err := tx.Insert(&block); err != nil {
			tx.Rollback() // 事务回滚
			return fmt.Errorf("tx.Insert: %s",err.Error())
		}
	}
	// 检查区块是否分叉
	if scanner.forkCheck(&block) {
		data, _ := json.Marshal(info)
		scanner.log("分叉！", string(data))
		tx.Commit()  // 即使分叉了，也要把保存区块的事务提交
		scanner.fork = true // 发生分叉
		return errors.New("fork check")  // 返回错误，让上层处理并重启区块扫描
	}
	// 解析区块内数据，读取内部的 “transactions” 交易信息，分析得出各种合约事件
	blockNumber := scanner.hexToTen(info.BlockNumber).Int64()
	scanner.log(
		"scan block start ==>", "number:", blockNumber, "hash:", info.BlockHash)
	scanner.client.TransactionHandler(info.Txs)
	scanner.log("scan block finish \n=================")
	if err = scanner.client.InsertTxsToDB(blockNumber,tx,info.Txs); err != nil {
		tx.Rollback()  // 事务回滚
		return fmt.Errorf("client.InsertTxsToDB: %s",err.Error())
	}
	return tx.Commit() // 提交事务
}
