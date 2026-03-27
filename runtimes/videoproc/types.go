package videoproc

import "tools/runtimes/videoproc/frame"

// 音频
type Audio struct {
	Src        string `json:"src"` // 需要替换的音频地址
	videoAudio string // 视频分离出来的音频
}

type Output struct {
	Mirror int    `json:"mirror"` // 视频镜像
	Audio  *Audio `json:"audio"`  // 替换的音频
}

type Maker struct {
	Srcs   []string     `json:"srcs" binding:"required"` // 源视频
	Frame  *frame.Frame `json:"frame"`                   // 帧处理
	Output *Output      `json:"output"`                  // 使用ffmpeg处理的

	tempdir string `json:"-"`
}
