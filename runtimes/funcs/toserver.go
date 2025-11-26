package funcs

import (
	"encoding/json"
	"time"
	"fmt"
	"tools/runtimes/aess"
)

type TokenStruct struct{
	Uuid string `json:"uuid" form:"uuid"`
	Exp   int64  `json:"exp" form:"exp"`
	Nonce string `json:"nonce" form:"nonce"`
}
func ServerHeader(version string, versionCode int) map[string]string{
	token := TokenStruct{
		Uuid: Uuid(),
		Exp: time.Now().Add(time.Minute * 5).Unix(),
		Nonce: NewNonce(),
	}

	header := map[string]string{
		"version": version,
		"version_code": fmt.Sprintf("%d",versionCode),
	}
	if tokenByte, err := json.Marshal(token); err == nil{
		tokenStr := aess.AesEncryptCBC(string(tokenByte))
		if tokenStr != ""{
			header["token"] 	= tokenStr
		}
	}
	return header
}
