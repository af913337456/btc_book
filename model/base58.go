package model

import (
	"bytes"
	"math/big"
)

// 编码表 b58Alphabet
var b58table = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// base58 编码，input 是字节流
func Base58Encode(input []byte) []byte {
	var result []byte
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(b58table)))
	zero := big.NewInt(0)
	mod := &big.Int{}
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58table[mod.Int64()])
	}
	ReverseBytes(result)
	for _, b := range input {
		if b == 0x00 {
			result = append([]byte{b58table[0]}, result...)
		} else {
			break
		}
	}
	return result
}

func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0

	for _, b := range input {
		if b != b58table[0] {
			break
		}

		zeroBytes++
	}

	payload := input[zeroBytes:]
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58table, b)
		result.Mul(result, big.NewInt(int64(len(b58table))))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}