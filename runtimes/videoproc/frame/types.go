package frame

// 视频帧,抽帧/补帧, 锐化,对比度,亮度，模糊
type Frame struct {
	Skipping            bool    `json:"skipping"`              // 是否抽帧
	RandomSkipping      bool    `json:"random_skipping"`       // 是否随机抽帧
	FrameNumberSkipping int     `json:"frame_number_skipping"` // 间隔N帧抽一帧
	Interpolation       bool    `json:"interpolation"`         // 是否补帧
	RandomInterpolation bool    `json:"random_interpolation"`  // 是否随机补帧
	FrameInterpolation  int     `json:"frame_interpolation"`   // 间隔N帧补一帧
	Clearer             bool    `json:"clearer"`               // 变清晰,需要realesrgan-ncnn-vulkan
	Sharpen             float64 `json:"sharpen"`               // 锐化
}
