package ffmpeg

import (
	"encoding/json"
	"errors"
	"strconv"
)

type AudioInfo struct {
	Duration   int64 // 毫秒
	Bitrate    string
	SampleRate string
	Channels   string
	Codec      string
}

// 获取音频信息
func GetAudioInfo(src string) (*AudioInfo, error) {
	str, _, err := RunFfporbe(
		true,
		"-protocol_whitelist", "file,http,https,tcp,tls",
		"-v",
		"quiet",
		"-print_format", "json",
		"-show_entries", "format=duration,bit_rate",
		"-show_entries", "stream=codec_name,sample_rate,channels",
		"-analyzeduration", "1000000",
		"-probesize", "1000000",
		"-i",
		src,
	)
	if err != nil {
		return nil, err
	}

	var result struct {
		Streams []struct {
			CodecName  string `json:"codec_name"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
		} `json:"streams"`

		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}

	err = json.Unmarshal([]byte(str), &result)
	if err != nil {
		return nil, err
	}

	if len(result.Streams) == 0 {
		return nil, errors.New("no audio stream")
	}

	durationFloat, _ := strconv.ParseFloat(result.Format.Duration, 64)

	info := &AudioInfo{
		Duration:   int64(durationFloat * 1000),
		Bitrate:    result.Format.BitRate,
		SampleRate: result.Streams[0].SampleRate,
		Channels:   strconv.Itoa(result.Streams[0].Channels),
		Codec:      result.Streams[0].CodecName,
	}

	return info, nil
}

// 从视频中分离出音频
func SeparateAudio(src, audioSrc string) error {
	if _, _, err := RunFfmpeg(
		true,
		"-i",
		src,
		"-vn",
		"-c:a",
		"libmp3lame",
		"-b:a",
		"192k",
		audioSrc,
	); err != nil {
		return err
	}
	return nil
}
