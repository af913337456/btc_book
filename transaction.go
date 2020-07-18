package btc_book

import (
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/btcsuite/btcd/chaincfg"
	"qiniupkg.com/x/errors.v7"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/txscript"
	"encoding/hex"
	"bytes"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btc_book/model"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2020/1/1
*/

// https://www.blockchain.com/btctest/tx/110212bcc0cb55a69b60924edeb83c2e68ee1d7c426f6cb7f73c4a99d191311d
// 110212bcc0cb55a69b60924edeb83c2e68ee1d7c426f6cb7f73c4a99d191311d
// 查看浏览器：https://www.blockchain.com/btctest/tx/b005c81d851bd2d9d6017de2d9b83d29674df641d75b1839006e676b874d3d54

// 发送测试网络的比特币交易，value 是转账的数值，1 代表的是 0.000000001 BTC
type OpReturnDataObj struct {
	Data string
}
func SendTestNet_BTCNormalTransaction(senderPrivateKey,toAddress string,value int64,opReturn *OpReturnDataObj) error {

	targetTransactionValue := btcutil.Amount(value)
	blockCypherApiTestNet := "btc/test3"

	// 根据发送者私钥得出它的其他信息，比如地址
	wallet := CreateWalletFromPrivateKey(senderPrivateKey,&chaincfg.TestNet3Params)
	if wallet == nil {
		return errors.New("invalid private key") // 恢复钱包失败，非法私钥
	}
	// 1. ----------- 准备交易的输入，即UTXO -----------
	// 使用 三方平台 blockCypher 提供的 api 来获取发送者的 utxo 列表
	senderAddress := wallet.GetBtcAddress(true)
	utxoList,err := getUTXOListFromBlockCypherAPI(senderAddress,blockCypherApiTestNet)
	if err != nil {
		return err
	}

	// 2. ----------- 根据交易的数值选择好输入的条数，utxoList 是候选 UTXO 列表 -----------
	tx := wire.NewMsgTx(wire.TxVersion) // 定义一个交易对象
	var (
		totalUTXOValue btcutil.Amount
		changeValue btcutil.Amount
	)
	// SpendSize 是 BTC 建议的数值，用于参与手续费计算
	// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
	SpendSize := 1 + 73 + 1 + 33
	for _, utxo := range utxoList.Txrefs {
		totalUTXOValue += btcutil.Amount(utxo.BtcValue) // 统计可用的 UTXO 数值
		hash := &chainhash.Hash{}
		if err := chainhash.Decode(hash,utxo.TxHash);err != nil {
			panic(fmt.Errorf("构造 hash 错误 %s",err.Error()))
		}

		// 以上一条交易的 hash，n 参数来构建出本次交易的输入，即 UTXO
		preUTXO  := wire.OutPoint{Hash:*hash,Index:uint32(utxo.TxOutputN)}
		oneInput := wire.NewTxIn(&preUTXO, nil, nil)

		tx.AddTxIn(oneInput) // 添加到要使用的 UTXO 列表

		// 根据交易的数据量大小计算手续费
		txSize := tx.SerializeSize() + SpendSize * len(tx.TxIn)
		reqFee := btcutil.Amount(txSize * 10)

		// 候选 UTXO 总额减去需要的手续费 再 和目标转账值比较
		if totalUTXOValue - reqFee < targetTransactionValue {
			// 还没满足要转的数值，就继续循环
			continue
		}
		// 3. ----------- 给自己找零，计算好找零金额值 -----------
		changeValue = totalUTXOValue - targetTransactionValue - reqFee
		break // 如果已经满足了转账数值，那么就要跳出循环了，不再累加 UTXO
	}

	// 4. ----------- 构建交易的输出 -----------
	// 因为我们要做的一般的，给个人钱包地址转账，所以我们使用源码中提供的 PayToAddrScript 函数即可
	toPubkeyHash := getAddressPubkeyHash(toAddress)
	if toPubkeyHash == nil {
		return errors.New("invalid receiver address") // 非法钱包地址
	}
	toAddressPubKeyHashObj, err := btcutil.NewAddressPubKeyHash(toPubkeyHash,&chaincfg.TestNet3Params)
	if err != nil {
		return err
	}
	// 下面的 toAddressLockScript 是锁定脚本，PayToAddrScript 函数是源码提供的
	toAddressLockScript, err := txscript.PayToAddrScript(toAddressPubKeyHashObj)
	if err != nil {
		return err
	}
	// receiverOutput 对应收款者的输出
	receiverOutput := &wire.TxOut{PkScript: toAddressLockScript, Value: int64(targetTransactionValue)}
	tx.AddTxOut(receiverOutput) // 添加进交易结构里面

	// 如果有设置 opReturn ，那么组装一个 opReturn 的输出到交易里面
	if opReturn != nil {
		// NullDataScript 函数是库提供的
		nullDataScript,err := txscript.NullDataScript([]byte(opReturn.Data))
		if err != nil {
			return err
		}
		opreturnOutput := &wire.TxOut{PkScript: nullDataScript, Value: 0}
		tx.AddTxOut(opreturnOutput) // 添加进交易结构里面
	}

	var senderAddressLockScript []byte
	if changeValue > 0 { // 数值大于 0 ，那么我们需要给自己 sender 找零
		// 首先计算自己的锁定脚本值，计算方式和上面的一样
		senderPubkeyHash := getAddressPubkeyHash(senderAddress)
		senderAddressPubKeyHashObj, err := btcutil.NewAddressPubKeyHash(senderPubkeyHash,&chaincfg.TestNet3Params)
		if err != nil {
			return err
		}
		// 下面的 toAddressLockScript 是锁定脚本，PayToAddrScript 函数是源码提供的
		senderAddressLockScript, err = txscript.PayToAddrScript(senderAddressPubKeyHashObj)
		if err != nil {
			return err
		}
		// senderOutput 对应发送者的找零输出
		senderOutput := &wire.TxOut{PkScript: senderAddressLockScript, Value: int64(changeValue)}
		tx.AddTxOut(senderOutput) // 添加进交易结构里面
	}
	// 对每条输入，使用发送者私钥生成签名脚本，以标明这是发送者可用的
	btcecPrivateKey := (btcec.PrivateKey)(wallet.PrivateKey)
	txInSize := len(tx.TxIn)
	for i := 0; i<txInSize; i++ {
		sigScript, err :=
			txscript.SignatureScript( // 签名脚本生成函数由源码提供
				tx,
				i,
				senderAddressLockScript,
				txscript.SigHashAll,
				&btcecPrivateKey,
				true)
		if err != nil {
			return err
		}
		tx.TxIn[i].SignatureScript = sigScript // 赋值签名脚本
	}

	// 5. ----------- 发送交易 -----------
	// 首先得出交易的 hash 格式的值，发送交易给节点，发的是 hash
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err = tx.Serialize(buf); err != nil {
		return err
	}
	txHex := hex.EncodeToString(buf.Bytes())
	// 发送交易数据到节点
	txHash,err := sendRawTransactionHexToNode_BlockCypherAPI(txHex,blockCypherApiTestNet)
	if err != nil {
		return err
	}
	fmt.Println("交易的 hash 是:",txHash)
	return nil
}

// 获取一个地址的 pubkeyHash
func getAddressPubkeyHash(address string) []byte {
	pubKeyHash := model.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	return pubKeyHash
}

const (
	blockCypherAccessToken = "60f554d6fbad4dcd8ba1cb42cd1d3178"
)

// 接收请求接口成功后返回的结构体
type UTXORet struct {
	Balance      int64 `json:"balance"`
	FinalBalance int64 `json:"final_balance"`
	Txrefs []struct{
		TxHash      string `json:"tx_hash"`
		BlockHeight int64  `json:"block_height"`
		TxInputN    int64  `json:"tx_input_n"`
		TxOutputN   int64  `json:"tx_output_n"`
		BtcValue    int64  `json:"value"`
		Spent       bool   `json:"spent"`
	} `json:"txrefs"`
}
// 使用第三方平台 blockCypher 提供的 api 来获取 UTXO
func getUTXOListFromBlockCypherAPI(address,netType string) (*UTXORet,error) {
	number := 1000 // 限制最多获取这个地址 1000 条 UTXO
	url :=
		fmt.Sprintf(
		"https://api.blockcypher.com/v1/%s/addrs/%s?unspentOnly=true&limit=%d" +
			"&includeScript=false&includeConfidence=false",
		netType,address,number) // 组装好 blockCypher 网站 api 链接
	url = url + "&" + blockCypherAccessToken // 携带上 blockCypher 的 access token
	req,err := http.NewRequest("GET",url,strings.NewReader(""))
	if err != nil {
		return nil,err
	}
	resp,err := (&http.Client{}).Do(req) // 开始请求
	if err != nil {
		return nil,err
	}
	data,err := ioutil.ReadAll(resp.Body) // 读取数据
	defer resp.Body.Close()
	if err != nil {
		return nil,nil
	}
	// 解析数据
	utxoList := UTXORet{}
	if err = json.Unmarshal(data,&utxoList);err != nil {
		return nil,err
	}
	return &utxoList,nil
}

type SendTxRet struct {
	Tx struct{
		Hash string `json:"hash"`
	} `json:"tx"`
}
// 使用第三方平台 blockCypher 提供的 api 来发送交易
func sendRawTransactionHexToNode_BlockCypherAPI(txHex,netType string) (string,error) {
	// 下面是发送交易的 api 链接
	url := fmt.Sprintf("https://api.blockcypher.com/v1/%s/txs/push",netType)
	url = url + "?" + blockCypherAccessToken // 携带上 blockCypher 的 access token
	// {“tx”:$TXHEX}
	jsonStr := fmt.Sprintf("{\"tx\":\"%s\"}",txHex) // 构造参数，把交易 hash 放置进去
	req, err := http.NewRequest("POST", url,strings.NewReader(jsonStr))
	if err != nil {
		return "",err
	}
	req.Header.Set("Content-Type", "application/json")
	resp,err := (&http.Client{}).Do(req)
	if err != nil {
		return "",err
	}
	data,err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "",nil
	}
	fmt.Println("请求发送交易结果：\n",string(data))
	ret := SendTxRet{}
	if err = json.Unmarshal(data,&ret);err != nil {
		return "",nil
	}
	return ret.Tx.Hash,nil
}

































