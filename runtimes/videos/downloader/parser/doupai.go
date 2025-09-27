package parser

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type douPai struct {
}

func (d douPai) parseShareUrl(shareUrl string, transport *http.Transport) (*VideoParseInfo, error) {
	urlInfo, err := url.Parse(shareUrl)
	if err != nil {
		return nil, errors.New("parse share url fail")
	}
	if len(urlInfo.Query()["id"]) <= 0 {
		return nil, errors.New("can not parse video id from share url")
	}

	return d.parseVideoID(urlInfo.Query()["id"][0], transport)
}

func (d douPai) parseVideoID(videoId string, transport *http.Transport) (*VideoParseInfo, error) {
	reqUrl := fmt.Sprintf("https://v2.doupai.cc/topic/%s.json", videoId)
	headers := map[string]string{
		HttpHeaderUserAgent: DefaultUserAgent,
	}

	client := resty.New()
	if transport != nil {
		client.SetTransport(transport)
	}
	res, err := client.R().
		SetHeaders(headers).
		Get(reqUrl)
	if err != nil {
		return nil, err
	}

	data := gjson.GetBytes(res.Body(), "data")
	parseInfo := &VideoParseInfo{
		Title:    data.Get("name").String(),
		VideoUrl: data.Get("videoUrl").String(),
		CoverUrl: data.Get("imageUrl").String(),
		Platform: "doupai",
	}
	parseInfo.Author.Uid = data.Get("userId.id").String()
	parseInfo.Author.Name = data.Get("userId.name").String()
	parseInfo.Author.Avatar = data.Get("userId.avatar").String()

	return parseInfo, nil
}
