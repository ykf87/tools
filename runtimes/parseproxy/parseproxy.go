package parseproxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// 解析代理到配置
func ParseProxy(input string) (*Proxy, error) {

	// 1. URL 解析（vmess:// vless:// socks://）
	if strings.Contains(input, "://") {
		return parseFromURL(input)
	}

	// 2. JSON（很多机场用）
	if strings.HasPrefix(input, "{") {
		return parseFromJSON(input)
	}

	return nil, fmt.Errorf("unsupported proxy format")
}

func ExtractHostPort(input string) string {

	// 正则：匹配 host:port 或 host
	re := regexp.MustCompile(`([a-zA-Z0-9\.\-]+)(?::(\d+))?`)

	matches := re.FindAllStringSubmatch(input, -1)

	for _, m := range matches {

		host := m[1]

		// 过滤掉明显不是 host 的（比如 "socks" "http"）
		if !isValidHost(host) {
			continue
		}

		var port int
		if len(m) > 2 && m[2] != "" {
			port, _ = strconv.Atoi(m[2])
		}

		return fmt.Sprintf("%s:%d", host, port)
	}

	return ""
}

func isValidHost(host string) bool {

	// 1. IP
	if net.ParseIP(host) != nil {
		return true
	}

	// 2. 域名（简单判断）
	if regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(host) {
		return true
	}

	return false
}
func parseFromURL(raw string) (*Proxy, error) {
	u, err := url.Parse(raw)
	if err != nil {
		ipport := ExtractHostPort(raw)
		tmp := strings.ReplaceAll(raw, ipport, "")
		tmp = strings.Replace(tmp, "@", "", -1)
		raw = tmp + "@" + ipport
		u, err = url.Parse(raw)
		if err != nil {
			return nil, err
		}
	}
	switch u.Scheme {
	case "vmess":
		return parseVMess(raw)
	case "vless":
		return parseVLess(u)
	case "socks", "socks5":
		return parseSocks(raw)
	case "trojan":
		return parseTrojan(u)
	case "http", "https":
		return parseHTTP(u)
	case "ss":
		return parseSS(raw)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}

func parseSocks(raw string) (*Proxy, error) {
	// 先标准解析
	u, err := url.Parse(raw)
	if err == nil && u.Host != "" {

		port, _ := strconv.Atoi(u.Port())

		p := &Proxy{
			Type:   "socks5",
			Server: u.Hostname(),
			Port:   port,
		}

		if u.User != nil {
			p.Username = u.User.Username()
			p.Password, _ = u.User.Password()
		}

		return p, nil
	}

	// fallback：手动解析（处理脏格式）
	raw = strings.TrimPrefix(raw, "socks://")
	raw = strings.TrimPrefix(raw, "socks5://")

	parts := strings.Split(raw, "@")

	var hostPart, authPart string

	if len(parts) == 2 {
		// 判断哪边是 host
		if isHostPort(parts[0]) {
			hostPart = parts[0]
			authPart = parts[1]
		} else {
			hostPart = parts[1]
			authPart = parts[0]
		}
	} else {
		hostPart = raw
	}

	host, port := splitHostPort(hostPart)

	p := &Proxy{
		Type:   "socks5",
		Server: host,
		Port:   port,
	}

	// 解析用户名密码
	if authPart != "" {
		up := strings.Split(authPart, ":")
		if len(up) >= 1 {
			p.Username = up[0]
		}
		if len(up) >= 2 {
			p.Password = up[1]
		}
	}

	return p, nil
}
func isHostPort(s string) bool {
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		return false
	}
	return host != ""
}
func splitHostPort(s string) (string, int) {

	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return s, 0
	}

	port, _ := strconv.Atoi(portStr)
	return host, port
}

func parseVLess(u *url.URL) (*Proxy, error) {
	port, _ := strconv.Atoi(u.Port())
	q := u.Query()

	p := &Proxy{
		Type:   "vless",
		Server: u.Hostname(),
		Port:   port,
		UUID:   u.User.Username(),
	}

	p.Network = q.Get("type")
	p.Security = q.Get("security")
	p.TLS = p.Security == "tls"
	p.SNI = q.Get("sni")

	// ws
	if p.Network == "ws" {
		p.WSOpts = &WSOptions{
			Path: q.Get("path"),
		}
		// p.WSPath = q.Get("path")
	}

	return p, nil
}
func parseVMess(raw string) (*Proxy, error) {

	// vmess://base64(json)
	data := strings.TrimPrefix(raw, "vmess://")

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(decoded, &m); err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(fmt.Sprint(m["port"]))

	p := &Proxy{
		Type:     "vmess",
		Server:   fmt.Sprint(m["add"]),
		Port:     port,
		UUID:     fmt.Sprint(m["id"]),
		AlterID:  toInt(m["aid"]),
		Network:  fmt.Sprint(m["net"]),
		Security: fmt.Sprint(m["scy"]),
	}

	// ws
	if p.Network == "ws" {
		p.WSOpts = &WSOptions{
			Path: fmt.Sprint(m["path"]),
		}
	}

	return p, nil
}

func parseSS(raw string) (*Proxy, error) {

	raw = strings.TrimPrefix(raw, "ss://")

	var name string

	// 处理 #name
	if idx := strings.Index(raw, "#"); idx != -1 {
		name, _ = url.QueryUnescape(raw[idx+1:])
		raw = raw[:idx]
	}

	var method, password, host string
	var port int

	// 判断是否是 base64 整体
	if !strings.Contains(raw, "@") {
		// 整体 base64
		decoded, err := decodeBase64(raw)
		if err != nil {
			return nil, err
		}
		raw = decoded
	}

	// 拆分 userinfo 和 host
	parts := strings.Split(raw, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss format")
	}

	userInfo := parts[0]
	hostPart := parts[1]

	// userInfo 可能是 base64
	if strings.Contains(userInfo, ":") {
		// 明文
		mp := strings.SplitN(userInfo, ":", 2)
		method = mp[0]
		password = mp[1]
	} else {
		// base64(method:password)
		decoded, err := decodeBase64(userInfo)
		if err != nil {
			return nil, err
		}
		mp := strings.SplitN(decoded, ":", 2)
		if len(mp) != 2 {
			return nil, fmt.Errorf("invalid userinfo")
		}
		method = mp[0]
		password = mp[1]
	}

	// 解析 host:port
	h, p := splitHostPort(hostPart)
	host = h
	port = p

	return &Proxy{
		Name:     name,
		Type:     "ss",
		Server:   host,
		Port:     port,
		Cipher:   method,
		Password: password,
	}, nil
}

func decodeBase64(s string) (string, error) {

	// URL safe
	if decoded, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return string(decoded), nil
	}

	// 标准
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		return string(decoded), nil
	}

	return "", fmt.Errorf("invalid base64")
}

func toInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	case json.Number:
		i, _ := val.Int64()
		return int(i)
	default:
		return 0
	}
}

func parseTrojan(u *url.URL) (*Proxy, error) {
	port, _ := strconv.Atoi(u.Port())

	p := &Proxy{
		Type:     "trojan",
		Server:   u.Hostname(),
		Port:     port,
		Password: u.User.Username(),
		TLS:      true,
	}

	q := u.Query()
	p.SNI = q.Get("sni")

	return p, nil
}

func parseHTTP(u *url.URL) (*Proxy, error) {
	portStr := u.Port()
	port := 0

	if portStr == "" {
		if u.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	} else {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		port = p
	}

	p := &Proxy{
		Type:   "http",
		Server: u.Hostname(),
		Port:   port,
	}

	if u.User != nil {
		p.Username = u.User.Username()
		p.Password, _ = u.User.Password()
	}

	return p, nil
}

func parseFromJSON(raw string) (*Proxy, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}

	// 可以根据 type 分发
	t := fmt.Sprint(m["type"])

	switch t {
	case "vmess":
		return mapToProxy(m), nil
	case "vless":
		return mapToProxy(m), nil
	}

	return nil, fmt.Errorf("unsupported json proxy")
}

func mapToProxy(m map[string]any) *Proxy {

	p := &Proxy{
		Type: getString(m, "type"),
		Raw:  m,
	}

	// ===== 基础字段 =====
	p.Name = getString(m, "name", "ps")
	p.Server = getString(m, "server", "add", "host")
	p.Port = getInt(m, "port")

	p.Username = getString(m, "username", "user")
	p.Password = getString(m, "password", "pass")

	// ===== 核心协议字段 =====
	p.UUID = getString(m, "uuid", "id")
	p.AlterID = getInt(m, "alterId", "aid")

	p.Security = getString(m, "security", "scy", "cipher")
	p.Network = getString(m, "network", "net", "type")

	// ===== TLS =====
	sec := strings.ToLower(getString(m, "security"))
	if sec == "tls" {
		p.TLS = true
	}
	if getBool(m, "tls") {
		p.TLS = true
	}

	p.SNI = getString(m, "sni", "servername")

	// ===== WS =====
	p.WSOpts = &WSOptions{
		Path: getString(m, "path"),
	}
	// p.WSPath = getString(m, "path")
	if headers, ok := m["headers"].(map[string]any); ok {
		p.WSOpts.Headers = map[string]string{}
		// p.WSHeaders = map[string]string{}
		for k, v := range headers {
			p.WSOpts.Headers[k] = fmt.Sprint(v)
		}
	}

	// ===== gRPC =====
	p.GRPCOpts = &GRPCOptions{
		ServiceName: getString(m, "serviceName"),
	}
	// p.ServiceName = getString(m, "serviceName")

	return p
}

func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	return ""
}

func getInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch val := v.(type) {
			case int:
				return val
			case float64:
				return int(val)
			case string:
				i, _ := strconv.Atoi(val)
				return i
			}
		}
	}
	return 0
}

func getBool(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch val := v.(type) {
			case bool:
				return val
			case string:
				return val == "true"
			}
		}
	}
	return false
}
