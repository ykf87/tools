package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/logs"

	"github.com/tidwall/gjson"
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
	Transfers  []string
	ConfMd5    string
	server     *core.Instance
	Guard      bool // 是否是进程守护,也就是和程序一起运行,不执行关闭操作
}

var proxysMap sync.Map

// 获取配置是否在监听
func IsRuning(configStr string) *ProxyConfig {
	confMd5 := funcs.Md5String(configStr)
	if p, ok := proxysMap.Load(confMd5); ok { // 如果已经启动则直接返回
		pc := p.(*ProxyConfig)
		if pc.server == nil {
			proxysMap.Delete(confMd5)
		} else {
			return pc
		}
	}
	return nil
}

// 获取用于启动的结构体,内部会解析出outbound
func Client(configStr, addr string, port int) (*ProxyConfig, error) {
	confMd5 := funcs.Md5String(configStr)
	if p, ok := proxysMap.Load(confMd5); ok { // 如果已经启动则直接返回
		pc := p.(*ProxyConfig)
		if pc.server == nil {
			proxysMap.Delete(confMd5)
		} else {
			return pc, nil
		}
	}

	cfg, err := ParseProxy(configStr)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("Run error")
	}

	if addr == "" {
		addr = "0.0.0.0"
	}

	if port == 0 { // 如果未指定端口,自动获取可用端口
		p, err := funcs.FreePort()
		if err != nil {
			return nil, err
		}
		port = p
	} else { // 如果指定了端口,则检查端口是否已被占用
		if ok := funcs.IsPortAvailable(port); !ok {
			return nil, fmt.Errorf(i18n.T("Port %d is already in use", port))
		}
	}

	cfg.ListenAddr = addr
	cfg.ListenPort = port
	cfg.ConfMd5 = confMd5
	return cfg, nil
}

// 服务重启,守护进程也可用重启
func (this *ProxyConfig) Restart() error {
	if this.server != nil {
		if err := this.server.Close(); err != nil {
			return err
		}
	}
	proxysMap.Delete(this.ConfMd5)

	_, err := this.Run(this.Guard, this.Transfers...)
	return err
}

/**
启动代理监听. guard 为是否是守护代理,守护代理无法被停止
transfers - 桥接代理的配置
*/
func (this *ProxyConfig) Run(guard bool, transfers ...string) (*ProxyConfig, error) {
	this.Guard = guard

	configJSON, err := this.GenerateXrayConfig(transfers...)
	if err != nil {
		return nil, err
	}

	server, err := core.New(configJSON)
	if err != nil {
		return nil, err
	}

	if err := server.Start(); err != nil {
		return nil, err
	}
	this.server = server
	this.Transfers = transfers

	proxysMap.Store(this.ConfMd5, this)

	return this, nil
}

// 关闭代理
func (this *ProxyConfig) Close() error {
	if this.Guard == true {
		return fmt.Errorf(i18n.T("The daemon agent cannot be shut down"))
	}
	if this.server != nil {
		if err := this.server.Close(); err != nil {
			return err
		}
	}
	proxysMap.Delete(this.ConfMd5)
	return nil
}

// 手动构造xray的启动配置
func (this *ProxyConfig) GenerateXrayConfig(transfers ...string) (*core.Config, error) {
	inbound, err := BuildInbound(this.ListenAddr, this.ListenPort)
	if err != nil {
		return nil, err
	}

	outbound, err := this.GetOutbound()
	if err != nil {
		return nil, err
	}

	var outs []*Outbound

	// 转接的代理配置
	for _, v := range transfers {
		cli, err := Client(v, "", 0)
		if err != nil {
			continue
		}
		pcs, err := cli.Run(false)
		if err == nil {
			o, err := pcs.GetOutbound()
			if err == nil {
				outs = append(outs, o)
			}
		}

	}
	outs = append(outs, outbound)

	configMap := map[string]any{
		"inbounds":  []any{inbound},
		"outbounds": outs,
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

// 获取监听的地址
func (this *ProxyConfig) Listened() string {
	if this.ListenAddr != "" && this.ListenPort > 100 {
		return fmt.Sprintf("http://%s:%d", this.ListenAddr, this.ListenPort)
	}
	return ""
}

// 获取ip所在国家iso
func GetLocal(configStr string, transfers ...string) (string, error) {
	confMd5 := funcs.Md5String(configStr)
	var pc *ProxyConfig
	pcm, ok := proxysMap.Load(confMd5)
	if !ok {
		// var port int
		// var err error
		// for {
		// 	port, err = funcs.FreePort()
		// 	if err == nil {
		// 		break
		// 	}
		// }
		cli, err := Client(configStr, "", 0)
		if err != nil {
			return "", err
		}
		ppp, err := cli.Run(false, transfers...)
		if err != nil {
			return "", err
		}
		pc = ppp
	} else {
		pc = pcm.(*ProxyConfig)
	}

	if pc.server == nil || pc.ListenAddr == "" {
		return "", fmt.Errorf(i18n.T("The proxy is not enabled"))
	}

	proxyUrl := pc.Listened()
	// fmt.Println(proxyUrl, "------- use proxy", pc.ListenAddr, pc.ListenPort)

	urlStr := "https://api.btloader.com/country"
	var transport *http.Transport
	var client *http.Client

	if proxyUrl != "" {
		proxyURL, err := url.Parse(proxyUrl)
		if err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			client = &http.Client{
				Transport: transport,
			}
		} else {
			return "", err
		}
	}
	if client == nil {
		client = &http.Client{}
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logs.Error(err.Error())
		return "", err
	}
	resp, err := client.Do(req)

	if err != nil {
		logs.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf(i18n.T("Return code error"))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err.Error())
		return "", err
	}

	gs := gjson.ParseBytes(body).Map()
	if gs["country"].Exists() {
		return strings.ToLower(gs["country"].String()), nil
	}
	return "", fmt.Errorf(i18n.T("Unable to query country"))
}
