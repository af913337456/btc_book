package rpc

import (
	"github.com/btcsuite/btcd/rpcclient"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/wire"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
*/

type BTCRPCClient struct {
	NodeUrl  string            // 代表节点的url链接
	client  *rpcclient.Client  // 代表rpc客户端句柄实例
}

// 实例化一个 BTCRPCClient 指针对象
func NewBTCRPCHttpClient(nodeUrl,user,password string) *BTCRPCClient {
	connCfg := &rpcclient.ConnConfig {
		Host:         nodeUrl,
		User:         user,
		Pass:         password,
		HTTPPostMode: true, // true 代表只运行 HTTP Post 的访问模式
		DisableTLS:   true,
		// DisableTLS：
		// 		如果RPC服务开启了 https 的访问模式，那么建议始终使用TLS，
		// 		即设置该值为 false，否则用户名和密码将以明文形式通过网络发送。
		//		如果没使用 http，就设置该值为 true
	}
	// 当我们指定 rpc 使用 HTTP 模式的时候，下面实例化 client 的时候，ntfnHandlers 参数必须要设置为 nil 空值
	rpcClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		// 初始化失败，终结程序，并将错误信息显示到控制台中
		errInfo := fmt.Errorf("初始化 rpc client 失败%s",err.Error()).Error()
		panic(errInfo)
	}
	return &BTCRPCClient{
		NodeUrl: nodeUrl,
		client : rpcClient,
	}
}

// 实例化一个 BTCRPCClient 指针对象
func NewBTCRPCSocketClient(nodeUrl,user,password string) *BTCRPCClient {
	//certHomeDir := btcutil.AppDataDir("btcwallet", false)
	//certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	connCfg := &rpcclient.ConnConfig{
		Host:         nodeUrl,
		Endpoint:     "ws", // websocket 的连接方式，这里固定设置为 ws 字符串
		User:         user,
		Pass:         password,
		Certificates: nil, // 如果节点服务开启了 https 的模式，那么这里要配置好对应的证书文件
	}
	handlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txs []*btcutil.Tx) {
			// OnFilteredBlockConnected 当一个区块被添加到最长链的时候，
			// 我们可以在这个回调函数中接收到事件
		},
	}
	rpcClient, err := rpcclient.New(connCfg, &handlers)
	if err != nil {
		// 初始化失败，终结程序，并将错误信息显示到控制台中
		errInfo := fmt.Errorf("初始化 rpc client 失败%s",err.Error()).Error()
		panic(errInfo)
	}
	return &BTCRPCClient{
		NodeUrl: nodeUrl,
		client : rpcClient,
	}
}

func (rpc *BTCRPCClient) Stop() { // 允许外部调用该函数停止 rpc 服务
	rpc.client.Shutdown()
}

// Go 语言语法中，大写字母开头的变量或者函数（方法）才能够被外部引用
// 小写字母的变量或函数（方法）只能内部调用
// GetRpc 函数（方法）是为了方便外部能够获取 client *rpcclient.Client，以便进行访问
func (rpc *BTCRPCClient) GetRpc() *rpcclient.Client {
	if rpc.client == nil {
		return nil
	}
	return rpc.client
}



