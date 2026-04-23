package proxys

import (
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/db/proxys"
	"tools/runtimes/parseproxy"
	"tools/runtimes/response"
	"tools/runtimes/services"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func Subscription(c *gin.Context) {
	suburl := fmt.Sprint(config.SERVERDOMAIN, "subscription")

	proxys, err := services.GerProxySub(suburl)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, gin.H{
		"url":  suburl,
		"list": proxys,
	}, "success")
}

// // 构建clash使用的yaml
// type Config struct {
// 	MixedPort int          `yaml:"mixed-port"`
// 	AllowLan  bool         `yaml:"allow-lan"`
// 	Mode      string       `yaml:"mode"`
// 	LogLevel  string       `yaml:"log-level"`
// 	Dns       Dns          `yaml:"dns"`
// 	Proxies   []Proxy      `yaml:"proxies"`
// 	Groups    []ProxyGroup `yaml:"proxy-groups"`
// 	Rules     []string     `yaml:"rules"`
// }

// type Dns struct {
// 	Enable         bool           `yaml:"enable"`
// 	Listen         string         `yaml:"listen"`
// 	EnhancedMode   string         `yaml:"enhanced-mode"`
// 	FakeIpRange    string         `yaml:"fake-ip-range"`
// 	Nameserver     []string       `yaml:"nameserver"`
// 	Fallback       []string       `yaml:"fallback"`
// 	FallbackFilter map[string]any `yaml:"fallback-filter"`
// }

// type Proxy struct {
// 	Name     string `yaml:"name"`
// 	Type     string `yaml:"type"`
// 	Server   string `yaml:"server"`
// 	Port     int    `yaml:"port"`
// 	Cipher   string `yaml:"cipher,omitempty"`
// 	Password string `yaml:"password,omitempty"`
// 	Username string `yaml:"username,omitempty"`
// }

// type ProxyGroup struct {
// 	Name    string   `yaml:"name"`
// 	Type    string   `yaml:"type"`
// 	Proxies []string `yaml:"proxies"`
// }

// func parseSS(raw string) (*Proxy, error) {

// 	u, err := url.Parse(raw)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if u.Scheme != "ss" {
// 		return nil, fmt.Errorf("not ss link")
// 	}

// 	// 1. 解析 userinfo（base64）
// 	userInfo := u.User.String()

// 	decoded, err := base64.StdEncoding.DecodeString(userInfo)
// 	if err != nil {
// 		// 有些是 URL-safe base64
// 		decoded, err = base64.RawURLEncoding.DecodeString(userInfo)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	parts := strings.SplitN(string(decoded), ":", 2)
// 	if len(parts) != 2 {
// 		return nil, fmt.Errorf("invalid ss userinfo")
// 	}

// 	method := parts[0]
// 	password := parts[1]

// 	// 2. host + port
// 	host := u.Hostname()
// 	port, _ := strconv.Atoi(u.Port())

// 	// 3. name（# 后）
// 	name, _ := url.QueryUnescape(u.Fragment)

// 	return &Proxy{
// 		Name:     name,
// 		Type:     "ss",
// 		Server:   host,
// 		Port:     port,
// 		Cipher:   method,
// 		Password: password,
// 	}, nil
// }

// func parseSocksSmart(raw, title string) (*Proxy, error) {

// 	raw = strings.TrimSpace(raw)

// 	// 统一 scheme
// 	raw = strings.Replace(raw, "socks5://", "socks://", 1)

// 	// 👉 关键：修复错误格式 host:port@user:pass
// 	if strings.Count(raw, "@") == 1 {
// 		parts := strings.Split(raw, "@")

// 		left := parts[0]  // 可能是 socks://host:port
// 		right := parts[1] // 可能是 user:pass

// 		// 判断左边是不是 host:port
// 		if strings.Count(left, ":") >= 2 && strings.HasPrefix(left, "socks://") {
// 			// 说明是错误写法，交换
// 			raw = "socks://" + right + "@" + strings.TrimPrefix(left, "socks://")
// 		}
// 	}

// 	u, err := url.Parse(raw)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if u.Scheme != "socks" {
// 		return nil, fmt.Errorf("not socks scheme")
// 	}

// 	host := u.Hostname()
// 	port, _ := strconv.Atoi(u.Port())

// 	var username, password string

// 	if u.User != nil {
// 		username = u.User.Username()
// 		password, _ = u.User.Password()
// 	}

// 	return &Proxy{
// 		Name:     title,
// 		Type:     "socks5",
// 		Server:   host,
// 		Port:     port,
// 		Username: username,
// 		Password: password,
// 	}, nil
// }

func Clash(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id 为空", nil)
		return
	}

	px := proxys.GetById(id)
	conf := px.GetConfig()
	transform := px.GetTransfer()

	cfg, err := parseproxy.MkClashYaml(parseproxy.ClashOpt{}, px.Name, conf, transform)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// outbound, err := parseSocksSmart(conf, px.Name)
	// if err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// trans, err := parseSS(transform)
	// if err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// cfg := buildConfig(*outbound, *trans)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// 关键 header（必须）
	c.Header("Content-Type", "text/yaml; charset=utf-8")

	// 可选（提高兼容性）
	c.Header("Content-Disposition", "inline; filename=config.yaml")

	c.Writer.Write(data)
}

// func buildConfig(outbound, trans Proxy) *Config {
// 	exclude := []string{
// 		"IP-CIDR,127.0.0.0/8,DIRECT",
// 		"IP-CIDR,192.168.0.0/16,DIRECT",
// 		"IP-CIDR,10.0.0.0/8,DIRECT",
// 		"DOMAIN-SUFFIX,local,DIRECT",
// 		"DOMAIN-SUFFIX,lan,DIRECT",
// 		"GEOIP,CN,DIRECT", //-------- 国内直连（核心）
// 		"MATCH,relay-chain",
// 	}
// 	// if ipnet, err := funcs.GetLocalIP(true); err == nil {
// 	// 	var iip []string
// 	// 	for idx, vls := range strings.Split(ipnet, ".") {
// 	// 		if idx > 1 {
// 	// 			iip = append(iip, "0")
// 	// 		} else {
// 	// 			iip = append(iip, vls)
// 	// 		}
// 	// 	}
// 	// 	exclude = append(exclude, fmt.Sprintf("IP-CIDR,%s/16,DIRECT", strings.Join(iip, ".")))
// 	// }
// 	return &Config{
// 		MixedPort: 7890,
// 		AllowLan:  false,
// 		Mode:      "rule",
// 		LogLevel:  "info",
// 		Dns: Dns{
// 			Enable:       true,
// 			Listen:       "0.0.0.0:1053",
// 			EnhancedMode: "fake-ip",
// 			FakeIpRange:  "198.18.0.1/16",
// 			Nameserver: []string{
// 				"223.5.5.5",
// 				"119.29.29.29",
// 			},
// 			Fallback: []string{
// 				"8.8.8.8",
// 				"1.1.1.1",
// 			},
// 			FallbackFilter: map[string]any{
// 				"geoip":      true,
// 				"geoip-code": "cn",
// 			},
// 		},
// 		Proxies: []Proxy{
// 			trans,
// 			outbound,
// 		},
// 		Groups: []ProxyGroup{
// 			{
// 				Name:    "relay-chain",
// 				Type:    "relay",
// 				Proxies: []string{trans.Name, outbound.Name},
// 			},
// 		},
// 		Rules: exclude,
// 	}
// }
