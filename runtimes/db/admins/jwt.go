package admins

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

type Claims struct {
	Id      int64  `json:"id"`
	Account string `json:"account"`
	Timer   int64  `json:"timer"`
	jwt.RegisteredClaims
}

func init() {
	uuid := funcs.Uuid()
	if uuid == "" {
		uuid = "Ykf~Jwt#Key"
	}
	jwtKey = []byte(uuid)
}

func (this *Admin) GenJwt() (string, error) {
	expirationTime := time.Now().Add(24 * 365 * time.Hour) // 有效期 24h
	claims := &Claims{
		Id:      this.Id,
		Account: this.Account,
		Timer:   this.Timer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ykf-tools", // 谁签发的
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// 通过jwt获取管理员
func GetAdminFromJwt(tokenStr string) (*Admin, error) {
	tokenStr = strings.ReplaceAll(tokenStr, "Bearer ", "")
	claims := &Claims{}
	// 解析并验证
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		fmt.Println(err, "--------", tokenStr)
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	adm := new(Admin)
	if err := db.DB.DB().Model(&Admin{}).Where("id = ?", claims.Id).First(adm).Error; err != nil {
		return nil, err
	}
	if adm.Id < 1 {
		return nil, errors.New(i18n.T("Account not found"))
	}

	if adm.Timer != claims.Timer {
		return nil, errors.New(i18n.T("Account logged in elsewhere"))
	}
	return adm, nil
}
