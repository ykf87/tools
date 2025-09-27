package parser

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/wujunwei928/parse-video/utils"
)

// ParseVideoShareUrlByRegexp 将分享链接信息, 进行正则表达式匹配到分享链接后, 再解析视频信息
func ParseVideoShareUrlByRegexp(shareMsg string, transport *http.Transport) (*VideoParseInfo, error) {
	videoShareUrl, err := utils.RegexpMatchUrlFromString(shareMsg)
	if err != nil {
		return nil, err
	}

	return ParseVideoShareUrl(videoShareUrl, transport)
}

// ParseVideoShareUrl 根据视频分享链接解析视频信息: 分享链接需是正常http链接
func ParseVideoShareUrl(shareUrl string, transport *http.Transport) (*VideoParseInfo, error) {
	// 根据分享url判断source
	source := ""
	for itemSource, itemSourceInfo := range videoSourceInfoMapping {
		for _, itemUrlDomain := range itemSourceInfo.VideoShareUrlDomain {
			if strings.Contains(shareUrl, itemUrlDomain) {
				source = itemSource
				break
			}
		}
		if len(source) > 0 {
			break
		}
	}

	// 没有找到对应source
	if len(source) <= 0 {
		return nil, fmt.Errorf("share url [%s] not have source config", shareUrl)
	}

	// 没有对应的视频链接解析方法
	urlParser := videoSourceInfoMapping[source].VideoShareUrlParser
	if urlParser == nil {
		return nil, fmt.Errorf("source %s has no video share url parser", source)
	}

	return urlParser.parseShareUrl(shareUrl, transport)
}

// ParseVideoId 根据视频id解析视频信息
func ParseVideoId(source, videoId string, transport *http.Transport) (*VideoParseInfo, error) {
	if len(videoId) <= 0 || len(source) <= 0 {
		return nil, errors.New("video id or source is empty")
	}

	idParser := videoSourceInfoMapping[source].VideoIdParser
	if idParser == nil {
		return nil, fmt.Errorf("source %s has no video id parser", source)
	}

	return idParser.parseVideoID(videoId, transport)
}

// BatchParseVideoId 根据视频id批量解析视频信息
func BatchParseVideoId(source string, videoIds []string, transport *http.Transport) (map[string]BatchParseItem, error) {
	if len(videoIds) <= 0 || len(source) <= 0 {
		return nil, errors.New("videos id or source is empty")
	}

	idParser := videoSourceInfoMapping[source].VideoIdParser
	if idParser == nil {
		return nil, fmt.Errorf("source %s has no video id parser", source)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	parseMap := make(map[string]BatchParseItem, len(videoIds))
	for _, v := range videoIds {
		wg.Add(1)
		videoId := v
		go func(videoId string) {
			defer wg.Done()

			parseInfo, parseErr := ParseVideoId(source, videoId, transport)
			mu.Lock()
			parseMap[videoId] = BatchParseItem{
				ParseInfo: parseInfo,
				Error:     parseErr,
			}
			mu.Unlock()
		}(videoId)
	}
	wg.Wait()

	return parseMap, nil
}
