package parseproxy

import "fmt"

// 构建clash使用的yaml
type Config struct {
	MixedPort   int           `yaml:"mixed-port"`
	AllowLan    bool          `yaml:"allow-lan"`
	Mode        string        `yaml:"mode"`
	LogLevel    string        `yaml:"log-level"`
	Ipv6        bool          `yaml:"ipv6"`
	UdpEnabled  bool          `yaml:"udp-enabled"`
	Udp         bool          `yaml:"udp"`
	TcpFastOpen bool          `yaml:"tcp-fast-open"`
	Proxies     []*Proxy      `yaml:"proxies"`
	Groups      []*ProxyGroup `yaml:"proxy-groups"`
	Rules       []string      `yaml:"rules"`
	Dns         Dns           `yaml:"dns"`
	Tun         *Tun          `yaml:"tun"`
}

type Tun struct {
	Enable              bool     `yaml:"enable"`
	Stack               string   `yaml:"stack"`
	DnsHijack           []string `yaml:"dns-hijack"`
	AutoRoute           bool     `yaml:"auto-route"`
	AutoDetectInterface bool     `yaml:"auto-detect-interface"`
	StrictRoute         bool     `yaml:"strict-route"`
}

type Dns struct {
	Enable       bool   `yaml:"enable"`
	Listen       string `yaml:"listen"`
	EnhancedMode string `yaml:"enhanced-mode"`
	// FakeIpRange    string         `yaml:"fake-ip-range"`
	Nameserver     []string       `yaml:"nameserver"`
	Fallback       []string       `yaml:"fallback"`
	FallbackFilter map[string]any `yaml:"fallback-filter"`
	// NameserverPolicy      map[string]string `yaml:"nameserver-policy"`
	ProxyServerNameserver []string `yaml:"proxy-server-nameserver"`
	Ipv6                  bool     `yaml:"ipv6"`
	RespectRules          bool     `yaml:"respect-rules"`
	CacheSize             int      `yaml:"cache-size"`
}
type ProxyGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies"`
}

type ClashOpt struct {
	Port     int // 端口
	AllowLan bool
	Rules    []string
}

func MkClashYaml(opt ClashOpt, name string, outbound string, traners ...string) (*Config, error) {
	cfg, err := mkcfg(opt)
	if err != nil {
		return nil, err
	}

	ob, err := ParseProxy(outbound)
	if err != nil {
		return nil, err
	}
	if ob.Name == "" {
		ob.Name = name
	}

	rbname := "PROXY"
	baseGroup := &ProxyGroup{
		Name:    rbname,
		Type:    "select",
		Proxies: []string{ob.Name},
	}

	var proxies []*Proxy
	if len(traners) > 0 {
		// var pgnames []string
		var prev *Proxy
		for k, v := range traners {
			tpx, err := ParseProxy(v)
			if err == nil {
				if tpx.Name == "" {
					tpx.Name = fmt.Sprintf("PROXY-%d", k+1)
				}
				if prev != nil {
					tpx.DialerProxy = prev.Name
				}
				proxies = append(proxies, tpx)
				// pgnames = append(pgnames, tpx.Name)
				prev = tpx
			}
		}
		// pgnames = append(pgnames, ob.Name)
		// rbname = "relay-chain"
		// baseGroup.Proxies = pgnames
		// baseGroup = &ProxyGroup{
		// 	Name:    rbname,
		// 	Type:    "select",
		// 	Proxies: pgnames,
		// }
	}
	if len(proxies) > 0 {
		ob.DialerProxy = proxies[len(proxies)-1].Name
	}

	proxies = append(proxies, ob)
	cfg.Proxies = proxies
	cfg.Groups = []*ProxyGroup{baseGroup}

	cfg.Rules = []string{
		"IP-CIDR,127.0.0.0/8,DIRECT",
		"IP-CIDR,192.168.0.0/16,DIRECT",
		"IP-CIDR,10.0.0.0/8,DIRECT",
		"IP-CIDR,172.16.0.0/12,DIRECT",
		// "DOMAIN-SUFFIX,local,DIRECT",
		// "DOMAIN-SUFFIX,lan,DIRECT",
		// "DOMAIN,dns.google," + rbname,
		// "DOMAIN,cloudflare-dns.com," + rbname,
		// "DOMAIN-SUFFIX,googleapis.com," + rbname,
		// "DOMAIN-SUFFIX,tiktok.com," + rbname,
		// "DOMAIN-SUFFIX,tiktokv.com," + rbname,
		// "DOMAIN-SUFFIX,byteoversea.com," + rbname,
		// "DOMAIN-SUFFIX,ibyteimg.com," + rbname,
		// "DOMAIN-SUFFIX,muscdn.com," + rbname,
		// "DOMAIN-SUFFIX,tiktokcdn.com," + rbname,
		// "DOMAIN-SUFFIX,google.com," + rbname,
		// "DOMAIN-SUFFIX,facebook.com," + rbname,
		// "DOMAIN-SUFFIX,instagram.com," + rbname,
		// "DOMAIN-KEYWORD,google," + rbname,
		"GEOIP,CN,DIRECT", //-------- 国内直连（核心）
		// "MATCH,relay-chain",
	}
	cfg.Rules = append(cfg.Rules, "MATCH,"+rbname)

	return cfg, nil
}

func mkcfg(opt ClashOpt) (*Config, error) {
	if opt.Port < 100 {
		// opt.Port, _ = funcs.FreePort()
		opt.Port = 7890
	}
	return &Config{
		MixedPort:   opt.Port,
		AllowLan:    false,
		Mode:        "rule",
		LogLevel:    "warning",
		Ipv6:        false,
		UdpEnabled:  true,
		Udp:         true,
		TcpFastOpen: true,
		Dns: Dns{
			Enable:       true,
			Listen:       "0.0.0.0:1053",
			EnhancedMode: "redir-host",
			Ipv6:         false,
			// FakeIpRange:  "198.18.0.1/16",
			Nameserver: []string{
				"1.1.1.1#PROXY",
				"1.0.0.1#PROXY",
				"8.8.8.8#PROXY",
			},
			FallbackFilter: map[string]any{
				"geoip":      true,
				"geoip-code": "CN",
			},
			ProxyServerNameserver: []string{
				"1.1.1.1",
				"8.8.8.8",
			},
			// Fallback: []string{
			// 	"https://8.8.8.8/dns-query#GLOBAL",
			// 	"https://8.8.4.4/dns-query#GLOBAL",
			// },
			RespectRules: false,
			CacheSize:    4096,
			// NameserverPolicy: map[string]string{
			// 	"geosite:geolocation-!cn": "https://1.1.1.1/dns-query",
			// },
		},
		Tun: &Tun{
			Enable: true,
			Stack:  "system",
			DnsHijack: []string{
				"any:53",
				// "tcp://any:53",
			},
			AutoRoute:           true,
			AutoDetectInterface: true,
			StrictRoute:         false,
		},
	}, nil
}
