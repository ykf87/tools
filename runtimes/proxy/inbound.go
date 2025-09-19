package proxy

import (
	"fmt"
	"tools/runtimes/funcs"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
)

type Inbound struct {
	Tag      string         `json:"tag"`
	Port     int            `json:"port"`
	Listen   string         `json:"listen"`
	Protocol string         `json:"protocol"`
	Sniffing map[string]any `json:"sniffing"`
	Settings map[string]any `json:"settings"`
}

var IntagIndex int

func BuildInbound(addr string, port int) (*Inbound, error) {
	if addr == "" {
		addr = "0.0.0.0"
	}
	var err error
	if port < 1000 {
		port, err = funcs.FreePort()
		if err != nil {
			return nil, err
		}
	}

	IntagIndex++
	inbound := &Inbound{
		Listen:   addr,
		Port:     port,
		Tag:      fmt.Sprintf("http%d", IntagIndex),
		Protocol: "http",
		Sniffing: map[string]any{"enabled": true, "destOverride": []string{"http", "tls"}, "routeOnly": false},
		Settings: map[string]any{"auth": "noauth", "udp": true, "allowTransparent": false},
	}
	return inbound, nil
}
