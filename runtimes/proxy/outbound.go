package proxy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"

	_ "github.com/xtls/xray-core/app/proxyman/outbound"
)

type Outbound struct {
	Protocol string         `json:"protocol"`
	Settings map[string]any `json:"settings"`
}

func ParseProxy(raw string) (*ProxyConfig, error) {
	switch {
	case strings.HasPrefix(raw, "socks://"):
		return parseSocks(raw)
	case strings.HasPrefix(raw, "ss://"):
		return parseSS(raw)
	case strings.HasPrefix(raw, "vmess://"):
		return parseVMess(raw)
	case strings.HasPrefix(raw, "vless://"):
		return parseVLess(raw)
	case strings.HasPrefix(raw, "trojan://"):
		return parseTrojan(raw)
	case strings.HasPrefix(raw, "http://"):
		return parseHttp(raw)
	case strings.HasPrefix(raw, "https://"):
		return parseHttps(raw)
	default:
		return nil, errors.New(i18n.T("unsupported proxy scheme"))
	}
}

func (this *ProxyConfig) GetOutbound() (*Outbound, error) {
	outbound := &Outbound{
		Protocol: this.Protocol,
		Settings: map[string]any{},
	}

	switch this.Protocol {
	case "ss", "shadowsocks":
		// SS 协议的 Settings 通常包含：
		// "servers": [{ "address": ..., "port": ..., "method": ..., "password": ... }]
		outbound.Protocol = "shadowsocks"
		servers := []map[string]any{
			{
				"address":  this.RemoteAddr,
				"port":     this.RemotePort,
				"method":   this.Security, // method 对应 Security 字段
				"password": this.Password,
			},
		}
		outbound.Settings["servers"] = servers
	case "socks":
		outbound.Settings["servers"] = []map[string]any{
			{"address": this.RemoteAddr, "port": this.RemotePort, "user": this.Username, "pass": this.Password},
		}
	case "vmess":
		outbound.Settings["vnext"] = []map[string]any{
			{
				"address": this.RemoteAddr,
				"port":    this.RemotePort,
				"users": []map[string]any{
					{"id": this.UUID, "alterId": 0, "security": this.Security},
				},
			},
		}
	case "vless":
		outbound.Settings["vnext"] = []map[string]any{
			{"address": this.RemoteAddr, "port": this.RemotePort, "users": []map[string]any{{"id": this.UUID}}},
		}
	case "trojan":
		outbound.Settings["servers"] = []map[string]any{
			{"address": this.RemoteAddr, "port": this.RemotePort, "password": this.Password},
		}
	case "http", "https":
		mmsfd := make(map[string]any)
		mmsfd["address"] = this.RemoteAddr
		mmsfd["port"] = this.RemotePort
		if this.Username != "" && this.Password != "" {
			mmsfd["users"] = []map[string]any{{
				"user": this.Username,
				"pass": this.Password,
			}}
		}
		outbound.Settings["servers"] = []map[string]any{
			mmsfd,
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", this.Protocol)
	}

	return outbound, nil
}

func parseQuery(q url.Values) map[string]string {
	m := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func parseSocks(socksURL string) (*ProxyConfig, error) {
	if !strings.HasPrefix(socksURL, "socks://") {
		return nil, errors.New(i18n.T("not a %s url", "socks"))
	}

	raw := strings.TrimPrefix(socksURL, "socks://")
	parts := strings.SplitN(raw, "?", 2)
	base64Part := parts[0]
	query := ""
	if len(parts) > 1 {
		query = parts[1]
	}

	var decoded string
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Part)
	if err != nil {
		// return nil, err
		decoded = base64Part
	} else {
		decoded = string(decodedBytes)
	}

	addr, port, un, pwd := getHostPortOrNamePwd(decoded)
	if addr == "" || port < 10 {
		return nil, fmt.Errorf("Wrong socks proxy")
	}

	cfg := &ProxyConfig{
		Protocol:   "socks",
		ListenAddr: "127.0.0.1", // 默认本地监听
		ListenPort: 0,           // 默认端口，可在调用时修改
		RemoteAddr: addr,
		RemotePort: port,
		Username:   un,
		Password:   pwd,
		Extra:      make(map[string]any),
	}

	if query != "" {
		vals, _ := url.ParseQuery(query)
		for k, v := range vals {
			cfg.Extra[k] = v[0]
		}
	}

	return cfg, nil
}

func parseSS(ssURL string) (*ProxyConfig, error) {
	if !strings.HasPrefix(ssURL, "ss://") {
		return nil, errors.New("not a ss:// url")
	}

	raw := strings.TrimPrefix(ssURL, "ss://")
	parts := strings.SplitN(raw, "#", 2) // 去掉 remark
	raw = parts[0]

	var userInfo string
	// 判断是否包含 @，如果包含可能是明文格式
	if strings.Contains(raw, "@") {
		userInfo = raw
	} else {
		// 否则认为是 Base64
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, err
		}
		userInfo = string(decoded)
	}

	upParts := strings.SplitN(userInfo, "@", 2)
	if len(upParts) != 2 {
		return nil, errors.New("invalid ss format: missing @")
	}

	// method:password
	methodPass := strings.SplitN(upParts[0], ":", 2)
	if len(methodPass) != 2 {
		if rrrs, err := funcs.Base64Decode(upParts[0]); err == nil {
			methodPass = strings.SplitN(rrrs, ":", 2)
		}
	}
	if len(methodPass) != 2 {
		return nil, errors.New("invalid ss user/pass format")
	}

	// host:port
	hostPort := strings.SplitN(upParts[1], ":", 2)
	if len(hostPort) != 2 {
		return nil, errors.New("invalid ss host/port format")
	}
	remotePort, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, err
	}

	cfg := &ProxyConfig{
		Protocol:   "ss",
		ListenAddr: "127.0.0.1",
		ListenPort: 0,
		RemoteAddr: hostPort[0],
		RemotePort: remotePort,
		Security:   methodPass[0],
		Password:   methodPass[1],
		Extra:      make(map[string]any),
	}
	return cfg, nil
}

func parseVMess(vmessURL string) (*ProxyConfig, error) {
	if !strings.HasPrefix(vmessURL, "vmess://") {
		return nil, errors.New("not a vmess:// url")
	}

	raw := strings.TrimPrefix(vmessURL, "vmess://")
	decodedBytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decodedBytes, &data); err != nil {
		return nil, err
	}

	var port int
	var addr string
	var uuid string
	var network string
	var path string
	var scy string

	port, _ = data["port"].(int)
	addr, _ = data["add"].(string)
	uuid, _ = data["id"].(string)
	network, _ = data["net"].(string)
	path, _ = data["path"].(string)
	scy, _ = data["scy"].(string)

	cfg := &ProxyConfig{
		Protocol:   "vmess",
		ListenAddr: "127.0.0.1",
		ListenPort: 0,
		RemoteAddr: addr,
		RemotePort: port,
		UUID:       uuid,
		Security:   scy,
		Network:    network,
		Path:       path,
		Extra:      make(map[string]any),
	}

	// 可以把 json 里其他字段放到 Extra
	for k, v := range data {
		if _, ok := cfg.Extra[k]; !ok {
			if s, ok := v.(string); ok {
				cfg.Extra[k] = s
			} else if i, ok := v.(int); ok {
				cfg.Extra[k] = i
			} else if b, ok := v.(bool); ok {
				cfg.Extra[k] = b
			}
		}
	}

	return cfg, nil
}

func parseVLess(vlessURL string) (*ProxyConfig, error) {
	if !strings.HasPrefix(vlessURL, "vless://") {
		return nil, errors.New("not a vless:// url")
	}

	raw := strings.TrimPrefix(vlessURL, "vless://")
	parts := strings.SplitN(raw, "?", 2)
	hostPortPart := parts[0]
	queryStr := ""
	if len(parts) > 1 {
		queryStr = parts[1]
	}

	upParts := strings.SplitN(hostPortPart, "@", 2)
	if len(upParts) != 2 {
		return nil, errors.New("invalid vless format")
	}

	uuid := upParts[0]
	hostPort := strings.SplitN(upParts[1], ":", 2)
	remotePort, _ := strconv.Atoi(hostPort[1])

	cfg := &ProxyConfig{
		Protocol:   "vless",
		ListenAddr: "127.0.0.1",
		ListenPort: 0,
		RemoteAddr: hostPort[0],
		RemotePort: remotePort,
		UUID:       uuid,
		Extra:      make(map[string]any),
	}

	if queryStr != "" {
		vals, _ := url.ParseQuery(queryStr)
		for k, v := range vals {
			cfg.Extra[k] = v[0]
		}
	}

	return cfg, nil
}

func parseTrojan(trojanURL string) (*ProxyConfig, error) {
	if !strings.HasPrefix(trojanURL, "trojan://") {
		return nil, errors.New("not a trojan:// url")
	}

	raw := strings.TrimPrefix(trojanURL, "trojan://")
	parts := strings.SplitN(raw, "?", 2)
	hostPortPart := parts[0]
	queryStr := ""
	if len(parts) > 1 {
		queryStr = parts[1]
	}

	upParts := strings.SplitN(hostPortPart, "@", 2)
	if len(upParts) != 2 {
		return nil, errors.New("invalid trojan format")
	}

	password := upParts[0]
	hostPort := strings.SplitN(upParts[1], ":", 2)
	remotePort, _ := strconv.Atoi(hostPort[1])

	cfg := &ProxyConfig{
		Protocol:   "trojan",
		ListenAddr: "127.0.0.1",
		ListenPort: 0,
		RemoteAddr: hostPort[0],
		RemotePort: remotePort,
		Password:   password,
		Extra:      make(map[string]any),
	}

	if queryStr != "" {
		vals, _ := url.ParseQuery(queryStr)
		for k, v := range vals {
			cfg.Extra[k] = v[0]
		}
	}

	return cfg, nil
}

func parseHttp(row string) (*ProxyConfig, error) {
	if !strings.HasPrefix(row, "http://") {
		return nil, errors.New("not a http url")
	}

	row = strings.TrimLeft(row, "http://")

	cfg := new(ProxyConfig)
	cfg.Protocol = "http"

	host, port, username, pwd := getHostPortOrNamePwd(strings.TrimLeft(row, "https://"))

	if host == "" || port < 10 {
		return nil, fmt.Errorf("Wrong http proxy")
	}
	cfg.RemoteAddr = host
	cfg.RemotePort = port
	cfg.Username = username
	cfg.Password = pwd

	if cfg.RemoteAddr == "" || cfg.RemotePort < 10 {
		return nil, fmt.Errorf("Wrong http proxy")
	}

	return cfg, nil
}

func parseHttps(row string) (*ProxyConfig, error) {
	if !strings.HasPrefix(row, "https://") {
		return nil, errors.New("not a https url")
	}

	cfg := new(ProxyConfig)
	cfg.Protocol = "https"

	host, port, username, pwd := getHostPortOrNamePwd(strings.TrimLeft(row, "https://"))

	if host == "" || port < 10 {
		return nil, fmt.Errorf("Wrong http proxy")
	}
	cfg.RemoteAddr = host
	cfg.RemotePort = port
	cfg.Username = username
	cfg.Password = pwd

	return cfg, nil
}

// 通过连接获取 host, port, username 和 password
func getHostPortOrNamePwd(str string) (string, int, string, string) {
	var remotePort int
	var remoteAddr, userName, passowrd string
	strs := strings.Split(str, "@")
	for _, v := range strs {
		if strings.Contains(v, ".") {
			hs := strings.Split(v, ":")
			for _, dsdf := range hs {
				if port, err := strconv.Atoi(dsdf); err == nil {
					remotePort = port
				} else if strings.Contains(dsdf, ".") {
					remoteAddr = dsdf
				}
			}
		} else if strings.Contains(v, ":") {
			sdfm := strings.Split(v, ":")
			if len(sdfm) == 2 {
				userName = sdfm[0]
				passowrd = sdfm[1]
			}
		}
	}
	return remoteAddr, remotePort, userName, passowrd
}
