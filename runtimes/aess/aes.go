package aess

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var key []byte
var iv []byte

func init() {
	key = []byte("!e+WmGnA^6i9gCOXCj+lA5@(rwyP(QBF")
	iv = []byte("BzVK05QLOhttvBuQ")
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncryptCBC(orig string) string {
	origData := []byte(orig)
	k := key

	block, _ := aes.NewCipher(k)

	blockSize := block.BlockSize()

	origData = PKCS7Padding(origData, blockSize)

	ivCopy := iv
	blockMode := cipher.NewCBCEncrypter(block, ivCopy[:blockSize])

	cryted := make([]byte, len(origData))

	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)//RawURLEncoding 将加密字符串转成url支持的格式,替换StdEncoding
}

func AesDecryptCBC(cryted string) string {
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)//RawURLEncoding 将加密字符串转成url支持的格式,替换StdEncoding
	if err != nil || len(crytedByte) <= 0 {
		return cryted
	}

	k := key
	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()

	ivCopy := iv
	blockMode := cipher.NewCBCDecrypter(block, ivCopy[:blockSize])
	orig := make([]byte, len(crytedByte))

	blockMode.CryptBlocks(orig, crytedByte)

	orig = PKCS7UnPadding(orig)
	return string(orig)
}
