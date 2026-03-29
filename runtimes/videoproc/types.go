package videoproc

import "tools/runtimes/imager"

// 音频
type Audio struct {
	Src        []string `json:"src"` // 需要替换的音频地址
	videoAudio []string `json:"-"`   // 视频分离出来的音频
}

type Factory struct {
	Mirror   int              `json:"mirror"`   // 视频镜像
	Crop     *imager.Crop     `json:"crop"`     // 裁剪
	Linear   *imager.Linear   `json:"linear"`   // 亮度和对比度跳转
	Rotation *imager.Rotation `json:"rotation"` // 旋转
	Resize   *imager.Resize   `json:"resize"`   // 调整大小
	Clearer  *imager.Clearer  `json:"clearer"`  // ai变清晰
}

type OriginVideos struct {
	Url string `json:"url"`
}
type OriginAudio struct {
	Url string `json:"url"`
}

type Maker struct {
	Srcs    []string `json:"srcs" binding:"required"` // 源视频
	Audios  []string `json:"audios"`                  // 需要替换的音频
	Factory *Factory `json:"factory"`                 // 操作方法

	srcs    []*OriginVideos
	audios  []*OriginAudio
	tempdir string
}
