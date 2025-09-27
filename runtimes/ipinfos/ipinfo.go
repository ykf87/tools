package ipinfos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"tools/runtimes/i18n"
	"tools/runtimes/logs"

	"github.com/tidwall/gjson"
)

type Ipinfo struct {
	Ip       string `json:"ip"`
	Lang     string `json:"lang"`
	Timezone string `json:"timezone"`
	Iso      string `json:"iso"`
}

type geters struct {
	Url  string            `json:"url"`
	Maps map[string]string `json:"maps"`
}

var GetMap = []*geters{
	&geters{
		Url: "https://ipapi.co/json",
		Maps: map[string]string{
			"ip":           "ip",
			"country_code": "iso",
			"timezone":     "timezone",
			"languages":    "lang",
		},
	},
	&geters{
		Url: "http://ip-api.com/json/",
		Maps: map[string]string{
			"query":       "ip",
			"countryCode": "iso",
			"timezone":    "timezone",
			"languages":   "lang",
		},
	},
	&geters{
		Url: "https://ipwho.is/",
		Maps: map[string]string{
			"ip":           "ip",
			"country_code": "iso",
			"timezone.id":  "timezone",
			"languages":    "lang",
		},
	},
	&geters{
		Url: "https://ifconfig.co/json",
		Maps: map[string]string{
			"ip":          "ip",
			"country_iso": "iso",
			"time_zone":   "timezone",
			"languages":   "lang",
		},
	},
	&geters{
		Url: "https://ipinfo.io/json",
		Maps: map[string]string{
			"ip":        "ip",
			"country":   "iso",
			"timezone":  "timezone",
			"languages": "lang",
		},
	},
	&geters{
		Url: "https://api.ipbase.com/v1/json/",
		Maps: map[string]string{
			"ip":           "ip",
			"country_code": "iso",
			"time_zone":    "timezone",
			"languages":    "lang",
		},
	},
	&geters{
		Url: "https://ident.me/json",
		Maps: map[string]string{
			"ip":        "ip",
			"cc":        "iso",
			"tz":        "timezone",
			"languages": "lang",
		},
	},
	&geters{
		Url: "https://api.btloader.com/country",
		Maps: map[string]string{
			"country": "iso",
		},
	},
}

func Get(proxy string) (*Ipinfo, error) {
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

	ifo, err := query(client)
	if err != nil {
		return nil, err
	}

	info := new(Ipinfo)

	if v, ok := ifo["ip"]; ok {
		info.Ip = v
	}
	if v, ok := ifo["iso"]; ok {
		info.Iso = strings.ToLower(v)
	}
	if v, ok := ifo["lang"]; ok {
		info.Lang = v
	}
	if v, ok := ifo["timezone"]; ok {
		info.Timezone = v
	}
	return info, nil
}

func query(client *http.Client) (map[string]string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var body []byte
	for _, v := range GetMap {
		req, err = http.NewRequest("GET", v.Url, nil)
		if err != nil {
			logs.Error(err.Error() + ":" + v.Url)
			continue
		}
		resp, err = client.Do(req)
		if err != nil {
			logs.Error(err.Error() + ":" + v.Url)
			continue
		}

		if resp.StatusCode != 200 {
			logs.Error(v.Url + fmt.Sprintf(": get code %d", resp.StatusCode))
			err = fmt.Errorf(i18n.T("Return code error"))
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		resp.Body.Close()

		info := make(map[string]string)
		gs := gjson.ParseBytes(body).Map()

		for remoteKey, myKey := range v.Maps {
			if strings.Contains(remoteKey, ".") {
				info[myKey] = gjson.GetBytes(body, remoteKey).String()
			} else {
				if gs[remoteKey].Exists() {
					info[myKey] = gs[remoteKey].String()
				}
			}

		}

		return info, nil
	}
	return nil, err
}
