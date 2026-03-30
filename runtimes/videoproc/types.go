package videoproc

import (
	"context"
	"tools/runtimes/imager"
)

const (
	IMGTYPE = "bmp"
	IMGDIR  = "frames"
)

type Factory struct {
	Mirror   int              `json:"mirror"`   // 视频镜像,1水平， 2垂直
	Crop     *imager.Crop     `json:"crop"`     // 裁剪
	Linear   *imager.Linear   `json:"linear"`   // 亮度和对比度跳转
	Rotation *imager.Rotation `json:"rotation"` // 旋转
	Resize   *imager.Resize   `json:"resize"`   // 调整大小
	Clearer  bool             `json:"clearer"`  // ai变清晰
}

type Audio struct {
	Url        string `json:"url"`
	Codec      string // aac
	SampleRate int    // 44100
	Channels   int    // 2
	Bitrate    int64

	Duration float64
	Volume   float64 // 音量
}
type Video struct {
	Url     string `json:"url"`
	Codec   string // h264
	Width   int
	Height  int
	PixFmt  string // yuv420p
	FPS     float64
	Bitrate int64

	Duration  float64
	Audio     *Audio
	ImgNumers int // 拆分成图片后的图片数量
}

type MediaInfo struct {
	Format   string // mp4
	Duration float64
	Size     int64

	Video *Video
	Audio *Audio
}

// 视频ffmprob解析的信息
type ffprobeOutput struct {
	Streams []struct {
		CodecType    string `json:"codec_type"`
		CodecName    string `json:"codec_name"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		PixFmt       string `json:"pix_fmt"`
		BitRate      string `json:"bit_rate"`
		SampleRate   string `json:"sample_rate"`
		Channels     int    `json:"channels"`
		AvgFrameRate string `json:"avg_frame_rate"`
		Duration     string `json:"duration"`
	} `json:"streams"`

	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
	} `json:"format"`
}

type AudioInpter struct {
	Url    string  `json:"url"`
	Volume float64 `json:"volume"` // 音量
}

type Maker struct {
	ctx       context.Context
	Srcs      []string     `json:"srcs" binding:"required"` // 源视频
	Audio     *AudioInpter `json:"audio"`                   // 需要替换的源音频,也是bgm
	AmixAudio *AudioInpter `json:"amix_aduio"`              // 混合音频
	Factory   *Factory     `json:"factory"`                 // 操作方法
	Width     int          `json:"width"`                   // 输出宽度
	Height    int          `json:"height"`                  // 输出高度

	srcs      []*MediaInfo
	audio     *MediaInfo // 最终替换到视频的音频
	tempdir   string     // 临时目录,包含分离的音频, 视频帧目录等
	framesDir string     // 视频帧存储目录
}
