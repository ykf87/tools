package requests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

type Client struct {
	httpClient *http.Client
}

type Config struct {
	Timeout time.Duration
	Proxy   string // 支持 http:// 和 socks5://
}

func New(cfg *Config) (*Client, error) {
	transport := &http.Transport{}

	if cfg != nil && cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, err
		}

		if proxyURL.Scheme == "socks5" {
			dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
			if err != nil {
				return nil, err
			}

			// 优先使用带 Context 的拨号器（如果可用）
			if ctxDialer, ok := dialer.(proxy.ContextDialer); ok {
				transport.DialContext = ctxDialer.DialContext
			} else {
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}
		} else {
			// HTTP/HTTPS 代理
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	timeout := 15 * time.Second
	if cfg != nil && cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &Client{httpClient: client}, nil
}

func (c *Client) Get(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("http error: " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (c *Client) Post(url string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("http error: " + resp.Status)
	}
	respBody, err := io.ReadAll(resp.Body)
	return respBody, err
}
