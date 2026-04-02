package videoproc

import (
	"encoding/json"
	"strconv"
	"strings"
	"tools/runtimes/ffmpeg"
)

func ProbeMedia(path string) (*MediaInfo, error) {
	out, _, err := ffmpeg.RunFfporbe(true,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	if err != nil {
		return nil, err
	}

	var data ffprobeOutput
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		return nil, err
	}

	media := &MediaInfo{}

	// format
	media.Format = data.Format.FormatName
	media.Duration, _ = strconv.ParseFloat(data.Format.Duration, 64)
	media.Size, _ = strconv.ParseInt(data.Format.Size, 10, 64)

	// streams
	for _, s := range data.Streams {
		switch s.CodecType {
		case "video":
			fps := parseFPS(s.AvgFrameRate)
			bitrate, _ := strconv.ParseInt(s.BitRate, 10, 64)
			duration, _ := strconv.ParseFloat(s.Duration, 64)

			media.Video = &Video{
				Codec:    s.CodecName,
				Width:    s.Width,
				Height:   s.Height,
				PixFmt:   s.PixFmt,
				FPS:      fps,
				Bitrate:  bitrate,
				Duration: duration,
				Url:      path,
			}

		case "audio":
			bitrate, _ := strconv.ParseInt(s.BitRate, 10, 64)
			sr, _ := strconv.Atoi(s.SampleRate)
			duration, _ := strconv.ParseFloat(s.Duration, 64)

			media.Audio = &Audio{
				Codec:      s.CodecName,
				SampleRate: sr,
				Channels:   s.Channels,
				Bitrate:    bitrate,
				Duration:   duration,
				Url:        path,
			}
		}
	}

	return media, nil
}

func parseFPS(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) != 2 {
		return 0
	}

	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)

	if den == 0 {
		return 0
	}
	return num / den
}
