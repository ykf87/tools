package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type ProxyConfig struct {
	Protocol   string
	ListenAddr string
	ListenPort int
	RemoteAddr string
	RemotePort int
	UUID       string
	Password   string
	Username   string
	Security   string
	Network    string
	Path       string
	Extra      map[string]any
}

func Run(configStr, addr string, port int) (string, error) {
	cfg, err := ParseProxy(configStr)
	if err != nil {
		return "", err
	}

	configJSON, err := GenerateXrayConfig(cfg, addr, port)
	if err != nil {
		return "", err
	}

	server, err := core.New(configJSON)

	if err != nil {
		return "", err
	}
	if err := server.Start(); err != nil {
		return "", err
	}
	return "", nil
}

func GenerateXrayConfig(proxy *ProxyConfig, addr string, port int) (*core.Config, error) {
	inbound, err := BuildInbound(addr, port)
	if err != nil {
		return nil, err
	}
	fmt.Println(addr, port)

	outbound := &Outbound{
		Protocol: proxy.Protocol,
		Settings: map[string]any{},
	}

	switch proxy.Protocol {
	case "ss", "shadowsocks":
		// SS 协议的 Settings 通常包含：
		// "servers": [{ "address": ..., "port": ..., "method": ..., "password": ... }]
		outbound.Protocol = "shadowsocks"
		servers := []map[string]any{
			{
				"address":  proxy.RemoteAddr,
				"port":     proxy.RemotePort,
				"method":   proxy.Security, // method 对应 Security 字段
				"password": proxy.Password,
			},
		}
		outbound.Settings["servers"] = servers
	case "socks":
		outbound.Settings["servers"] = []map[string]any{
			{"address": proxy.RemoteAddr, "port": proxy.RemotePort, "user": proxy.Username, "pass": proxy.Password},
		}
	case "vmess":
		outbound.Settings["vnext"] = []map[string]any{
			{
				"address": proxy.RemoteAddr,
				"port":    proxy.RemotePort,
				"users": []map[string]any{
					{"id": proxy.UUID, "alterId": 0, "security": proxy.Security},
				},
			},
		}
	case "vless":
		outbound.Settings["vnext"] = []map[string]any{
			{"address": proxy.RemoteAddr, "port": proxy.RemotePort, "users": []map[string]any{{"id": proxy.UUID}}},
		}
	case "trojan":
		outbound.Settings["servers"] = []map[string]any{
			{"address": proxy.RemoteAddr, "port": proxy.RemotePort, "password": proxy.Password},
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proxy.Protocol)
	}

	configMap := map[string]any{
		"inbounds":  []any{inbound},
		"outbounds": []any{outbound},
	}

	data, err := json.Marshal(configMap)
	if err != nil {
		return nil, err
	}

	cf, err := serial.ReaderDecoderByFormat["json"](strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	coreConfig, err := cf.Build()
	if err != nil {
		return nil, err
	}

	return coreConfig, nil
}
