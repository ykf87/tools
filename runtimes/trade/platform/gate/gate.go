package gate

import (
	"context"

	"github.com/gate/gateapi-go/v7"
)

type Gate struct {
	key    string
	secret string
	ctx    context.Context
	client *gateapi.APIClient
}

func Client(key, secret, proxy string) *Gate {
	g := &Gate{
		key:    key,
		secret: secret,
	}

	g.client = gateapi.NewAPIClient(gateapi.NewConfiguration())
	g.ctx = context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    key,
		Secret: secret,
	})

	return g
}
