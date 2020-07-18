package btc_book

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btc_book/model"
)

/**
    作者(Author): 林冠宏 / 指尖下的幽灵
    Created on : 2019/11/17
*/

// 比特币钱包结构体
type BtcWallet struct {
	PrivateKey ecdsa.PrivateKey     // 私钥
	PublicKey  *btcec.PublicKey     // 公钥
	chainType  *chaincfg.Params     // 链的类型参数，比如主网，测试网
}

// 根据链的类型创建钱包指针对象
func CreateBtcWallet(chainType *chaincfg.Params) *BtcWallet {
	private, public := newKeyPair()  // 使用椭圆曲线生成私钥和公钥
	wallet := BtcWallet{PrivateKey:private, PublicKey:public}
	wallet.chainType = chainType
	return &wallet
}

// 创建多签钱包
type MultiWallet struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
}
// n 就是N个私钥来解锁
// pubKeyStrs 是参与生成多签地址的公钥数组
// chainType 是指定要生成何种节点网络的多签地址，有主网的，测试网的 和 私人网
func CreateMultiWallet(n int,pubKeyStrs []string, chainType *chaincfg.Params) (*MultiWallet, error) {
	pubKeyList := make([]*btcutil.AddressPubKey, len(pubKeyStrs))
	// 在循环里面逐个添加参与生成多签地址的公钥
	for index, pubKeyItem := range pubKeyStrs {
		// DecodeAddress 把字符串类型的公钥解码成 address 对象
		addressObj, err := btcutil.DecodeAddress(pubKeyItem, chainType)
		if err != nil {
			return nil,fmt.Errorf("invalid pubkey %s: decode failed err %s", pubKeyItem, err.Error())
		}
		if !addressObj.IsForNet(chainType) { // 判断该公钥是否对应当前的链类型
			return nil,fmt.Errorf("invalid pubkey %s: not match chain type %s", pubKeyItem, err.Error())
		}
		switch pubKey := addressObj.(type) {
		case *btcutil.AddressPubKey:
			pubKeyList[index] = pubKey
			continue
		default:
			// 出现了不符合要求的地址数据，直接返回错误。比如传参的不是公钥
			return nil, fmt.Errorf("address contains invalid item")
		}
	}
	script, err := txscript.MultiSigScript(pubKeyList, n)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scirpt %s", err.Error())
	}
	address, err := btcutil.NewAddressScriptHash(script, chainType)
	if err != nil {
		return nil, err
	}
	return &MultiWallet{
		Address: address.EncodeAddress(), // 地址
		// 这里使用 EncodeToString 编码返回 16 进制 redeemScript
		RedeemScript: hex.EncodeToString(script),
	}, nil
}

// 使用椭圆曲线生成私钥和公钥
func newKeyPair() (ecdsa.PrivateKey, *btcec.PublicKey) {
	curve := elliptic.P256()
	// ecdsa 是椭圆曲线数字签名算法的英文简称
	// GenerateKey 函数帮助我们生成了私钥的二进制数据，封装好了返回
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err) // 如果有错误，那么我们中断程序
	}
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), private.D.Bytes())
	return *private, (*btcec.PublicKey)(pubKey)
}

// 导入 WIF 格式的私钥，恢复钱包
func CreateWalletFromPrivateKey(wifPrivateKey string,chainType *chaincfg.Params) *BtcWallet {
	wif, err := btcutil.DecodeWIF(wifPrivateKey)
	if err != nil {
		panic(err)
	}
	privKeyBytes := wif.PrivKey.Serialize()
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return &BtcWallet{
		chainType:chainType,
		PrivateKey:(ecdsa.PrivateKey)(*privKey),
		PublicKey:pubKey}
}

func (w *BtcWallet) GetWIFPrivateKey(compress bool) string {
	if w.PrivateKey.D == nil {
		return ""
	}
	var combine = []byte{}
	// 0x80 是 wif 的版本号
	if compress  {
		// 压缩版本
		combine = append([]byte{0x80},w.PrivateKey.D.Bytes()...)
		combine = append(combine,0x01)
	}else{
		combine = append([]byte{0x80},w.PrivateKey.D.Bytes()...)
	}
	checkCodeBytes := doubleSha256F(combine)
	combine = append(combine,checkCodeBytes[0:4]...)
	return string(model.Base58Encode(combine))
}

// 进行两次 sha256 哈希运算
func doubleSha256F(payload []byte) []byte {
	sha256H := sha256.New()
	sha256H.Reset()
	sha256H.Write(payload)
	hash1 := sha256H.Sum(nil)
	sha256H.Reset()
	sha256H.Write(hash1)
	return sha256H.Sum(nil)
}

func (w *BtcWallet) GetPubKeyHexStr(compress bool) string {
	if compress {
		// 压缩版
		return hex.EncodeToString(w.PublicKey.SerializeCompressed())
	}
	// 非压缩
	return hex.EncodeToString(w.PublicKey.SerializeUncompressed())
}

// 生成比特币地址
func (w *BtcWallet) GetBtcAddress(compress bool) string {
	if w.PublicKey == nil {
		return ""
	}
	var buf []byte
	if compress {
		buf = w.PublicKey.SerializeCompressed() // 压缩版
	}else {
		buf = w.PublicKey.SerializeUncompressed() // 非压缩版
	}
	if buf == nil {
		return ""
	}
	// Hash160 内部做了 sha256 和 ripemd160 操作
	pubKeyHash := btcutil.Hash160(buf)
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash,w.chainType)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	// addr.String 内部进行了 checkCode 的计算和 base58 编码
	return addr.String()
}

func (w *BtcWallet) GetBtcAddressPubkeyHash(compress bool) *btcutil.AddressPubKeyHash {
	var buf []byte
	if w.PublicKey == nil {
		return nil
	}
	if compress {
		buf = w.PublicKey.SerializeCompressed() // 压缩版
	}else {
		buf = w.PublicKey.SerializeUncompressed() // 非压缩版
	}
	// Hash160 内部做了 sha256 和 ripemd160 操作
	pubKeyHash := btcutil.Hash160(buf)
	addrPubkeyHash, err := btcutil.NewAddressPubKeyHash(pubKeyHash,w.chainType)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// addr.String 内部进行了 checkCode 的计算和 base58 编码
	return addrPubkeyHash
}

func (w *BtcWallet) GetBtcAddressObj(compress bool) btcutil.Address {
	var buf []byte
	if w.PublicKey == nil {
		return nil
	}
	if compress {
		buf = w.PublicKey.SerializeCompressed() // 压缩版
	}else {
		buf = w.PublicKey.SerializeUncompressed() // 非压缩版
	}
	_, addrPubkeyHash, _, err := txscript.ExtractPkScriptAddrs(buf,w.chainType)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// addr.String 内部进行了 checkCode 的计算和 base58 编码
	return addrPubkeyHash[0]
}

