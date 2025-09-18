package proxy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"

	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	// core "github.com/xtls/xray-core/core"
	// serials "github.com/xtls/xray-core/infra/conf/serial"
	// "github.com/xtls/xray-core/common/serial"
	// proxyss "github.com/xtls/xray-core/proxy/shadowsocks"
	// proxysocks "github.com/xtls/xray-core/proxy/socks"
	// proxytrojan "github.com/xtls/xray-core/proxy/trojan"
	// proxyvmess "github.com/xtls/xray-core/proxy/vmess"
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
	default:
		return nil, errors.New(i18n.T("unsupported proxy scheme"))
	}
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

	decodedBytes, err := base64.StdEncoding.DecodeString(base64Part)
	if err != nil {
		return nil, err
	}
	decoded := string(decodedBytes)

	// decoded 形如 "username:password@host:port"
	upParts := strings.SplitN(decoded, "@", 2)
	if len(upParts) != 2 {
		return nil, errors.New("invalid socks format")
	}

	userPass := strings.SplitN(upParts[0], ":", 2)
	hostPort := strings.SplitN(upParts[1], ":", 2)
	if len(userPass) != 2 || len(hostPort) != 2 {
		return nil, errors.New("invalid socks user/pass or host/port format")
	}

	remotePort, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, err
	}

	cfg := &ProxyConfig{
		Protocol:   "socks",
		ListenAddr: "127.0.0.1", // 默认本地监听
		ListenPort: 0,           // 默认端口，可在调用时修改
		RemoteAddr: hostPort[0],
		RemotePort: remotePort,
		Username:   userPass[0],
		Password:   userPass[1],
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

// func ParseOutbound(link string) (map[string]interface{}, error) {
// 	switch {
// 	case strings.HasPrefix(link, "ss://"):
// 		return parseSS(link)
// 	case strings.HasPrefix(link, "vmess://"):
// 		return parseVmess(link)
// 	case strings.HasPrefix(link, "trojan://"):
// 		return parseTrojan(link)
// 	// case strings.HasPrefix(link, "socks://"):
// 	// 	return parseSocks(link)
// 	default:
// 		return nil, fmt.Errorf("未知链接类型")
// 	}
// }

// // ---------- Shadowsocks ----------
// func parseSS(link string) (map[string]interface{}, error) {
// 	ssLink := strings.TrimPrefix(link, "ss://")
// 	if idx := strings.Index(ssLink, "#"); idx != -1 {
// 		ssLink = ssLink[:idx]
// 	}

// 	at := strings.LastIndex(ssLink, "@")
// 	if at == -1 {
// 		return nil, fmt.Errorf("ss 链接格式错误")
// 	}

// 	userInfoEnc := ssLink[:at]
// 	hostPort := ssLink[at+1:]

// 	userInfoBytes, err := base64.StdEncoding.DecodeString(userInfoEnc)
// 	if err != nil {
// 		return nil, fmt.Errorf("base64 解码失败: %v", err)
// 	}

// 	parts := strings.SplitN(string(userInfoBytes), ":", 2)
// 	if len(parts) != 2 {
// 		return nil, fmt.Errorf("user:pass 格式错误")
// 	}
// 	method := parts[0]
// 	password := parts[1]

// 	host, portStr, err := net.SplitHostPort(hostPort)
// 	if err != nil {
// 		return nil, fmt.Errorf("host:port 解析失败: %v", err)
// 	}

// 	port, _ := strconv.Atoi(portStr)

// 	return map[string]interface{}{
// 		"protocol": "shadowsocks",
// 		"settings": map[string]interface{}{
// 			"servers": []map[string]interface{}{
// 				{
// 					"address":  host,
// 					"port":     port,
// 					"method":   method,
// 					"password": password,
// 				},
// 			},
// 		},
// 		"streamSettings": map[string]interface{}{
// 			"network": "tcp",
// 		},
// 	}, nil

// 	// // 生成 JSON
// 	// ssJSON := map[string]interface{}{
// 	// 	"outbounds": []map[string]interface{}{
// 	// 		{
// 	// 			"protocol": "shadowsocks",
// 	// 			"settings": map[string]interface{}{
// 	// 				"servers": []map[string]interface{}{
// 	// 					{
// 	// 						"address":  host,
// 	// 						"port":     port,
// 	// 						"method":   method,
// 	// 						"password": password,
// 	// 					},
// 	// 				},
// 	// 			},
// 	// 			"streamSettings": map[string]interface{}{
// 	// 				"network": "tcp",
// 	// 			},
// 	// 		},
// 	// 	},
// 	// }

// 	// data, _ := json.Marshal(ssJSON)
// 	// reader := strings.NewReader(string(data))
// 	// cfg, err := serials.LoadJSONConfig(reader)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// return cfg, nil
// }

// // ---------- VMess ----------
// func parseVmess(link string) (map[string]interface{}, error) {
// 	vmessBase64 := strings.TrimPrefix(link, "vmess://")
// 	vmessJSONBytes, err := base64.StdEncoding.DecodeString(vmessBase64)
// 	if err != nil {
// 		return nil, fmt.Errorf("vmess base64 解码失败: %v", err)
// 	}

// 	resps := make(map[string]interface{})
// 	if err := json.Unmarshal(vmessJSONBytes, &resps); err != nil {
// 		return nil, err
// 	}
// 	return resps, nil

// 	// // vmessJSONBytes 是完整 JSON 配置片段
// 	// reader := strings.NewReader(string(vmessJSONBytes))
// 	// cfg, err := serials.LoadJSONConfig(reader)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// return cfg, nil
// }

// // ---------- Trojan ----------
// func parseTrojan(link string) (map[string]interface{}, error) {
// 	trojanLink := strings.TrimPrefix(link, "trojan://")
// 	if idx := strings.Index(trojanLink, "#"); idx != -1 {
// 		trojanLink = trojanLink[:idx]
// 	}

// 	at := strings.LastIndex(trojanLink, "@")
// 	if at == -1 {
// 		return nil, fmt.Errorf("trojan 链接格式错误")
// 	}
// 	password := trojanLink[:at]
// 	hostPort := trojanLink[at+1:]

// 	host, portStr, err := net.SplitHostPort(hostPort)
// 	if err != nil {
// 		return nil, fmt.Errorf("host:port 解析失败: %v", err)
// 	}
// 	port, _ := strconv.Atoi(portStr)

// 	return map[string]interface{}{
// 		"protocol": "trojan",
// 		"settings": map[string]interface{}{
// 			"servers": []map[string]interface{}{
// 				{
// 					"address":  host,
// 					"port":     port,
// 					"password": password,
// 				},
// 			},
// 		},
// 		"streamSettings": map[string]interface{}{
// 			"network": "tcp",
// 		},
// 	}, nil

// 	// trojanJSON := map[string]interface{}{
// 	// 	"outbounds": []map[string]interface{}{
// 	// 		{
// 	// 			"protocol": "trojan",
// 	// 			"settings": map[string]interface{}{
// 	// 				"servers": []map[string]interface{}{
// 	// 					{
// 	// 						"address":  host,
// 	// 						"port":     port,
// 	// 						"password": password,
// 	// 					},
// 	// 				},
// 	// 			},
// 	// 			"streamSettings": map[string]interface{}{
// 	// 				"network": "tcp",
// 	// 			},
// 	// 		},
// 	// 	},
// 	// }

// 	// data, _ := json.Marshal(trojanJSON)
// 	// reader := strings.NewReader(string(data))
// 	// cfg, err := serials.LoadJSONConfig(reader)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// return cfg, nil
// }

// // parseSocks 解析 socks:// 链接生成 Outbound
// // 支持格式：
// //   socks://host:port
// //   socks://user:pass@host:port
// func parseSocks(link string) (map[string]interface{}, error) {
// 	socksLink := strings.TrimPrefix(link, "socks://")
// 	// 去掉注释
// 	if idx := strings.Index(socksLink, "#"); idx != -1 {
// 		socksLink = socksLink[:idx]
// 	}

// 	var user, pass string
// 	var hostPort string

// 	if strings.Contains(socksLink, "@") {
// 		parts := strings.SplitN(socksLink, "@", 2)
// 		userPass := parts[0]
// 		hostPort = parts[1]
// 		up := strings.SplitN(userPass, ":", 2)
// 		if len(up) == 2 {
// 			user = up[0]
// 			pass = up[1]
// 		}
// 	} else {
// 		hostPort = socksLink
// 	}

// 	host, portStr, err := net.SplitHostPort(hostPort)
// 	if err != nil {
// 		return nil, fmt.Errorf("解析 host:port 失败: %v", err)
// 	}
// 	port, _ := strconv.Atoi(portStr)
// 	return map[string]interface{}{
// 		"protocol": "socks",
// 		"settings": map[string]interface{}{
// 			"servers": []map[string]interface{}{
// 				{
// 					"address": host,
// 					"port":    port,
// 					"user":    user,
// 					"pass":    pass,
// 				},
// 			},
// 		},
// 		"streamSettings": map[string]interface{}{
// 			"network": "tcp",
// 		},
// 	}, nil

// 	// // 构造 JSON 配置
// 	// socksJSON := map[string]interface{}{
// 	// 	"outbounds": []map[string]interface{}{
// 	// 		{
// 	// 			"protocol": "socks",
// 	// 			"settings": map[string]interface{}{
// 	// 				"servers": []map[string]interface{}{
// 	// 					{
// 	// 						"address": host,
// 	// 						"port":    port,
// 	// 						"user":    user,
// 	// 						"pass":    pass,
// 	// 					},
// 	// 				},
// 	// 			},
// 	// 			"streamSettings": map[string]interface{}{
// 	// 				"network": "tcp",
// 	// 			},
// 	// 		},
// 	// 	},
// 	// }

// 	// data, _ := json.Marshal(socksJSON)
// 	// reader := strings.NewReader(string(data))
// 	// cfg, err := serials.LoadJSONConfig(reader)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// return cfg, nil
// }
