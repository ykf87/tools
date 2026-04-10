package okx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	// goexv2 "github.com/nntaoli-project/goex/v2"
	// "github.com/nntaoli-project/goex/v2/httpcli"
	// "github.com/nntaoli-project/goex/v2/logger"
	// "github.com/nntaoli-project/goex/v2/okx/spot"
	// "github.com/nntaoli-project/goex/v2/options"
	// "github.com/nntaoli-project/goex/v2/util"
)

const (
	ENDPOINT = "https://www.okx.com"
)

type BaseResp struct {
	Code int             `json:"code,string"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type ErrorResponseData struct {
	OrdID   string `json:"ordId"`
	ClOrdId string `json:"clOrdId"`
	SCode   string `json:"sCode"`
	SMsg    string `json:"sMsg"`
}

type Okx struct {
	key      string
	secret   string
	password string
	client   *http.Client
}

var gHeader = make(map[string]string)

func Client(k, sec, pwd, proxy string) *Okx {
	var proxyClient *http.Transport

	if proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			proxyClient = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	ok := &Okx{
		key:      k,
		secret:   sec,
		password: pwd,
		client:   &http.Client{Transport: proxyClient},
	}

	// c := goexv2.OKx.Spot.NewPrvApi(
	// 	options.WithApiKey(k),
	// 	options.WithApiSecretKey(sec),
	// 	options.WithPassphrase(pwd))

	// ok.client = c
	return ok
}

func (o *Okx) doSignParam(httpMethod, apiUri, apiSecret, reqBody string) (signStr, timestamp string) {
	timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z") //iso time style
	payload := fmt.Sprintf("%s%s%s%s", timestamp, strings.ToUpper(httpMethod), apiUri, reqBody)
	signStr, _ = o.hmacSHA256Base64Sign(apiSecret, payload)
	return
}

func (o *Okx) hmacSHA256Base64Sign(secret, params string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(params))
	if err != nil {
		return "", err
	}
	signByte := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signByte), nil
}

func (o *Okx) hmacSHA256Sign(secret, params string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(params))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (o *Okx) valuesToJson(v url.Values) ([]byte, error) {
	paramMap := make(map[string]any)
	for k, vv := range v {
		if len(vv) == 1 {
			paramMap[k] = vv[0]
		} else {
			paramMap[k] = vv
		}
	}
	return json.Marshal(paramMap)
}

func (o *Okx) doAuthRequest(httpMethod, reqUrl string, params *url.Values, headers map[string]string) ([]byte, []byte, error) {
	var (
		reqBodyStr string
		reqUri     string
	)

	if http.MethodGet == httpMethod {
		reqUrl += "?" + params.Encode()
	}

	if http.MethodPost == httpMethod {
		// params.Set("tag", "86d4a3bf87bcBCDE")
		reqBody, _ := o.valuesToJson(*params)
		reqBodyStr = string(reqBody)
	}

	_url, _ := url.Parse(reqUrl)
	reqUri = _url.RequestURI()
	signStr, timestamp := o.doSignParam(httpMethod, reqUri, o.secret, reqBodyStr)

	headers = map[string]string{
		"Content-Type": "application/json; charset=UTF-8",
		//"Accept":               "application/json",
		"OK-ACCESS-KEY":        o.key,
		"OK-ACCESS-PASSPHRASE": o.password,
		"OK-ACCESS-SIGN":       signStr,
		"OK-ACCESS-TIMESTAMP":  timestamp}

	respBody, err := o.doRequest(httpMethod, reqUrl, reqBodyStr, headers)
	if err != nil {
		return nil, respBody, err
	}

	var baseResp BaseResp
	err = json.Unmarshal(respBody, &baseResp)
	if err != nil {
		return nil, respBody, err
	}

	if baseResp.Code != 0 {
		var errData []ErrorResponseData
		err = json.Unmarshal(baseResp.Data, &errData)
		if err != nil {
			// logger.Errorf("unmarshal error data error: %s", err.Error())
			return nil, respBody, errors.New(string(respBody))
		}
		if len(errData) > 0 {
			return nil, respBody, errors.New(errData[0].SMsg)
		}
		return nil, respBody, errors.New(baseResp.Msg)
	} // error response process

	return baseResp.Data, respBody, nil
}

func (o *Okx) doNoAuthRequest(httpMethod, reqUrl string, params *url.Values) ([]byte, []byte, error) {
	reqBody := ""
	if http.MethodGet == httpMethod {
		reqUrl += "?" + params.Encode()
	}

	responseBody, err := o.doRequest(httpMethod, reqUrl, reqBody, nil)
	if err != nil {
		return nil, responseBody, err
	}

	var baseResp BaseResp
	err = json.Unmarshal(responseBody, &baseResp)
	if err != nil {
		return responseBody, responseBody, err
	}

	if baseResp.Code == 0 {
		return baseResp.Data, responseBody, nil
	}

	return nil, responseBody, errors.New(baseResp.Msg)
}

func (o *Okx) doRequest(method, rqUrl string, reqBody string, headers map[string]string) (data []byte, err error) {
	reqTimeoutCtx, cancelFn := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancelFn()

	req, err := http.NewRequestWithContext(reqTimeoutCtx, method, fmt.Sprintf("%s%s", ENDPOINT, rqUrl), strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	//append global http header
	for k, v := range gHeader {
		req.Header.Set(k, v)
	}

	// if headers != nil {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// }

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
			// logger.Error("[DefaultHttpClient] close response body error:", err.Error())
		}
	}(resp.Body)

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %w", err)
	}

	if resp.StatusCode != 200 {
		return bodyData, errors.New(resp.Status)
	}

	return bodyData, nil
}
