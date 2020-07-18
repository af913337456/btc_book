package btc_book


/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2020/6/14
*/

import (
	"fmt"
	"errors"

	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btc_book/rpc"
	"github.com/btcsuite/btcd/btcjson"
)

func getUTXOListFromBitcoind(
	rpcClient *rpc.BTCRPCClient,
	pubkeyHash *btcutil.AddressPubKeyHash) ([]btcjson.ListUnspentResult,error) {

	// 比特币的 ListUnspent RPC接口，ListUnspentMinMaxAddresses
	list, err := rpcClient.GetRpc().ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{pubkeyHash})
	if err != nil {
		return nil, err
	}
	if list == nil || len(list) == 0 {
		return nil, errors.New("empty utxo list")
	}
	return list,nil
}
func SendLocalNode_BTCNormalTransaction(
	rpcClient *rpc.BTCRPCClient,
	senderPrivateKey,toAddress string,
	value int64,opReturn *OpReturnDataObj) error {

	netType := &chaincfg.RegressionNetParams
	targetTransactionValue := btcutil.Amount(value)

	// 根据发送者私钥得出它的其他信息，比如地址
	wallet := CreateWalletFromPrivateKey(senderPrivateKey,netType)
	if wallet == nil {
		return errors.New("invalid private key") // 恢复钱包失败，非法私钥
	}
	// 1. ----------- 准备交易的输入，即UTXO -----------
	// 使用 三方平台 blockCypher 提供的 api 来获取发送者的 utxo 列表
	senderAddressScriptHash := wallet.GetBtcAddressPubkeyHash(true)
	utxoList,err := getUTXOListFromBitcoind(rpcClient,senderAddressScriptHash)
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
	for _, utxo := range utxoList {
		totalUTXOValue += btcutil.Amount(utxo.Amount * 10e7) // 统计可用的 UTXO 数值
		hash := &chainhash.Hash{}
		if err := chainhash.Decode(hash,utxo.TxID);err != nil {
			panic(fmt.Errorf("构造 hash 错误 %s",err.Error()))
		}

		// 以上一条交易的 hash，n 参数来构建出本次交易的输入，即 UTXO
		preUTXO  := wire.OutPoint{Hash:*hash,Index: utxo.Vout}
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
	toAddressPubKeyHashObj, err := btcutil.NewAddressPubKeyHash(toPubkeyHash,netType)
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
		senderAddressPubKeyHashObj, err := btcutil.NewAddressPubKeyHash(senderAddressScriptHash.ScriptAddress(), netType)
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
	// 发送交易数据到节点
	txHash, err := rpcClient.GetRpc().SendRawTransaction(tx, false)
	if err != nil {
		panic(err)
	}
	fmt.Println("交易的 hash 是:",txHash)
	return nil
}
