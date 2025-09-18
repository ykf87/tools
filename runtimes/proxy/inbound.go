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

// func ParseInbound(proxyType, listenAddr string, listenPort uint16, tag string) map[string]interface{} {
// 	return map[string]interface{}{
// 		"tag":      tag,
// 		"protocol": "http",
// 		"listen":   listenAddr,
// 		"port":     listenPort,
// 		"settings": map[string]interface{}{},
// 		// "receiver_settings": map[string]interface{}{
// 		// 	"listen": listenAddr,
// 		// 	"port_range": map[string]interface{}{
// 		// 		"single": listenPort,
// 		// 	},
// 		// },
// 		// "proxy_settings": map[string]interface{}{
// 		// 	"protocol": proxyType,
// 		// 	"settings": map[string]interface{}{},
// 		// },
// 	}
// }

// func BuildHttpInbound(listenAddr string, listenPort uint16) (*serial.TypedMessage, error) {
// 	portRange := net.SinglePortRange(net.Port(listenPort))

// 	inbound := map[string]interface{}{
// 		"listen":   listenAddr,
// 		"port":     portRange,
// 		"protocol": "http",
// 		"settings": map[string]interface{}{
// 			"timeout": 0, // 可根据需要调整
// 		},
// 	}

// 	tm := serial.ToTypedMessage(inbound)
// 	if tm == nil {
// 		return nil, fmt.Errorf("生成 inbound 失败")
// 	}
// 	return tm, nil
// }

// func BuildTCPInbound(listenAddr string, listenPort uint16) (*serial.TypedMessage, error) {
// 	portRange := net.SinglePortRange(net.Port(listenPort))

// 	inbound := map[string]interface{}{
// 		"listen":   listenAddr,
// 		"port":     portRange,
// 		"protocol": "dokodemo-door", // TCP inbound，所有流量转发
// 		"settings": map[string]interface{}{
// 			"network":        "tcp",
// 			"followRedirect": true,
// 		},
// 	}

// 	tm := serial.ToTypedMessage(inbound)
// 	if tm == nil {
// 		return nil, fmt.Errorf("生成 inbound 失败")
// 	}
// 	return tm, nil
// }
