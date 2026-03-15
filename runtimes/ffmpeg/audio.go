package ffmpeg

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
)

type AudioInfo struct {
	Duration   int64
	Bitrate    string
	SampleRate string
	Channels   string
	Codec      string
}

func GetAudioInfo(src string) (*AudioInfo, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_entries", "format=duration,bit_rate",
		"-show_entries", "stream=codec_name,sample_rate,channels",
		"-analyzeduration", "1000000",
		"-probesize", "1000000",
		src,
	)

	out, err := cmd.Output()
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

	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Streams) == 0 {
		return nil, errors.New("no audio stream")
	}

	durationFloat, _ := strconv.ParseFloat(result.Format.Duration, 64)

	info := &AudioInfo{
		Duration:   int64(durationFloat),
		Bitrate:    result.Format.BitRate,
		SampleRate: result.Streams[0].SampleRate,
		Channels:   strconv.Itoa(result.Streams[0].Channels),
		Codec:      result.Streams[0].CodecName,
	}

	return info, nil
}
