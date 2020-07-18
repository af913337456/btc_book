package btc_book

import (
	"testing"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2019/11/17
*/

func TestCreateWallet(t *testing.T) {
	// 创建主链版本的钱包 MainNetParams
	btcWallet := CreateBtcWallet(&chaincfg.MainNetParams)
	t.Log("压缩版:",btcWallet.GetWIFPrivateKey(true))    // 输出压缩版 wif 私钥
	t.Log("非压缩版:",btcWallet.GetWIFPrivateKey(false)) // 非压缩版
	t.Log("压缩公钥:",btcWallet.GetPubKeyHexStr(true))
	t.Log("非压缩公钥:",btcWallet.GetPubKeyHexStr(false))
	t.Log("压缩地址:",btcWallet.GetBtcAddress(true))
	t.Log("非压缩地址:",btcWallet.GetBtcAddress(false))
}

func TestImportCreateWallet(t *testing.T) {
	btcWallet := CreateWalletFromPrivateKey(
		"cRgsH3pQMVhdux6HGzyMeRkoESfqtUNe5GhEvNw8Er7f4jTJwuoL",
		&chaincfg.RegressionNetParams)

	t.Log("压缩版:",btcWallet.GetWIFPrivateKey(true))    // 输出压缩版 wif 私钥
	t.Log("非压缩版:",btcWallet.GetWIFPrivateKey(false)) // 非压缩版
	t.Log("压缩公钥:",btcWallet.GetPubKeyHexStr(true))
	t.Log("非压缩公钥:",btcWallet.GetPubKeyHexStr(false))
	t.Log("压缩地址:",btcWallet.GetBtcAddress(true))
	t.Log("非压缩地址:",btcWallet.GetBtcAddress(false))

}

func TestBtcWallet_GetBtcAddress(t *testing.T) {
	addressStr := "03a622ed8e4b310a4f9065ef27cad4d77c510fed442c990f05ad0405fb5390c76b"
	address, err := btcutil.DecodeAddress(addressStr, &chaincfg.MainNetParams) // 把字符串类型的地址解码成 address 对象
	if err != nil {
		t.Log(err.Error())
	}else{
		t.Log(address.String())
	}
}

func Test_CreateMultiWallet(t *testing.T) {
	// pubkeys 是公钥数组，公钥的生成可以使用我们之前学习到的创建钱包函数
	pubkeys := []string{
		"03a622ed8e4b310a4f9065ef27cad4d77c510fed442c990f05ad0405fb5390c76b",
		"027c5412c9385dd30e03f50dae946e2742e589ae67fa8b89b1c41214a33008f226",
		"03e8ab2e68ad2c03df978c0233e69d21a14378bedd9d2404b73b685da173e918d7"}
	mutil,err := CreateMultiWallet(2,pubkeys,&chaincfg.MainNetParams)
	t.Log(err)
	if err == nil {
		t.Log(mutil.Address)
	}
}





