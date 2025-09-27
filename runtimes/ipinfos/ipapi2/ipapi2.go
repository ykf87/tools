package ipapi2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"tools/runtimes/i18n"

	"github.com/tidwall/gjson"
)

const GETURL = "http://ip-api.com/json/"

func Get(proxy string) (map[string]string, error) {
	if proxy == "" {
		return nil, fmt.Errorf("Empty")
	}

	var transport *http.Transport
	var client *http.Client

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client = &http.Client{
		Transport: transport,
	}
	if client == nil {
		client = &http.Client{}
	}
	req, err := http.NewRequest("GET", GETURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(i18n.T("Return code error"))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	gs := gjson.ParseBytes(body).Map()
	if gs["query"].Exists() {
		info["ip"] = gs["query"].String()
	}
	if gs["countryCode"].Exists() {
		info["iso"] = strings.ToLower(gs["countryCode"].String())
	}
	if gs["timezone"].Exists() {
		info["timezone"] = gs["timezone"].String()
	}
	if gs["languages"].Exists() {
		info["lang"] = gs["languages"].String()
	}

	return info, nil
}
