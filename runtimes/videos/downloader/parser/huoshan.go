package parser

import (
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type huoShan struct {
}

func (h huoShan) parseVideoID(videoId string, transport *http.Transport) (*VideoParseInfo, error) {
	reqUrl := "https://share.huoshan.com/api/item/info?item_id=" + videoId
	client := resty.New()
	if transport != nil {
		client.SetTransport(transport)
	}
	res, err := client.R().
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		Get(reqUrl)
	if err != nil {
		return nil, err
	}

	data := gjson.GetBytes(res.Body(), "data.item_info")
	videoUrl := data.Get("url").String()
	cover := data.Get("cover").String()

	parseRes := &VideoParseInfo{
		VideoUrl: videoUrl,
		CoverUrl: cover,
		Platform: "huoshan",
	}

	return parseRes, nil
}

func (h huoShan) parseShareUrl(shareUrl string, transport *http.Transport) (*VideoParseInfo, error) {
	client := resty.New()
	if transport != nil {
		client.SetTransport(transport)
	}
	// disable redirects in the HTTP client, get params before redirects
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, err := client.R().
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		Get(shareUrl)
	// 非 resty.ErrAutoRedirectDisabled 错误时，返回错误
	if !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		return nil, err
	}

	locationRes, err := res.RawResponse.Location()
	if err != nil {
		return nil, err
	}

	videoId := locationRes.Query().Get("item_id")
	if len(videoId) <= 0 {
		return nil, errors.New("parse video id from share url fail")
	}

	return h.parseVideoID(videoId, transport)
}
