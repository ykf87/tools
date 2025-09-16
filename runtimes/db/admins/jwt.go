package admins

import (
	"time"
	"tools/runtimes/funcs"

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
