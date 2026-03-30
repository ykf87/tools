package videoproc

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"tools/runtimes/ffmpeg"
)

func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// 音频干扰
func audioInterference(input *AudioInpter, output string) error {
	// 随机参数
	atempo := randRange(0.97, 1.05)
	pitch := randRange(0.98, 1.03)

	eq1Freq := randRange(300, 1200)
	eq1Gain := randRange(-3, 3)

	eq2Freq := randRange(2000, 6000)
	eq2Gain := randRange(-3, 3)

	if input.Volume <= 0 {
		input.Volume = randVolume()
	}

	// 构建 filter
	filter := fmt.Sprintf(
		"atempo=%.4f,asetrate=44100*%.4f,aresample=44100,"+
			"equalizer=f=%.0f:t=q:w=1:g=%.2f,"+
			"equalizer=f=%.0f:t=q:w=1:g=%.2f,"+
			"volume=%.3f",
		atempo,
		pitch,
		eq1Freq, eq1Gain,
		eq2Freq, eq2Gain,
		input.Volume,
	)

	if _, _, err := ffmpeg.RunFfmpeg(true,
		"-i", input.Url,
		"-filter_complex", filter,
		"-ar", "48000",
		"-ac", "2",
		"-b:a", "192k",
		output,
	); err != nil {
		return err
	}

	return nil
}

func randVolume() float64 {
	return 0.98 + rand.Float64()*(1.05-0.98)
}

// 合并音频（拼接 + 统一响度）
func audioMerge(inputs []*AudioInpter, output string) error {
	if len(inputs) == 0 {
		return errors.New("No Audios")
	}

	args := []string{}

	// 1️⃣ 输入
	for _, in := range inputs {
		args = append(args, "-i", in.Url)
	}

	// 2️⃣ filter
	var parts []string
	var labels []string

	for i := range inputs {
		label := fmt.Sprintf("a%d", i)

		parts = append(parts,
			fmt.Sprintf(
				"[%d:a]aresample=44100,"+
					"aformat=sample_fmts=fltp:channel_layouts=stereo,"+
					"asetpts=PTS-STARTPTS[%s]",
				i, label,
			),
		)

		labels = append(labels, fmt.Sprintf("[%s]", label))
	}

	// concat
	parts = append(parts,
		fmt.Sprintf("%sconcat=n=%d:v=0:a=1[cat]",
			strings.Join(labels, ""),
			len(inputs),
		),
	)

	// 👉 核心：统一响度 + 防爆音
	parts = append(parts,
		"[cat]anull[out]", //fmt.Sprintf("[cat]aformat=sample_rates=44100:channel_layouts=stereo,volume=%.4f,loudnorm=I=-16:TP=-1.5:LRA=11,alimiter=limit=0.95[out]", randVolume())
	)

	filter := strings.Join(parts, ";")

	// 3️⃣ 输出
	args = append(args,
		"-filter_complex", filter,
		"-map", "[out]",
		"-c:a", "aac",
		"-b:a", "192k",
		output,
	)

	_, _, err := ffmpeg.RunFfmpeg(true, args...)
	return err
}

// 混合音频
func audioAmix(inputs []*AudioInpter, output string) error {
	if len(inputs) == 0 {
		return errors.New("缺少音频")
	}

	args := []string{}

	// 1️⃣ 输入
	for _, in := range inputs {
		args = append(args, "-i", in.Url)
	}

	// 2️⃣ filter 构建
	var filterParts []string
	var labels []string

	for i, in := range inputs {
		label := fmt.Sprintf("a%d", i)

		// 默认音量
		volume := in.Volume
		if volume <= 0 {
			volume = 1.0
		}

		// ⚠️ 关键：加 aresample 防止不同音频不兼容
		filterParts = append(filterParts,
			fmt.Sprintf("[%d:a]aresample=44100,volume=%.3f[%s]", i, volume, label),
		)

		labels = append(labels, fmt.Sprintf("[%s]", label))
	}

	// 3️⃣ amix（防爆音）
	amix := fmt.Sprintf(
		"%samix=inputs=%d:duration=longest:dropout_transition=2",
		strings.Join(labels, ""),
		len(inputs),
	)

	filterParts = append(filterParts, amix)

	filter := strings.Join(filterParts, ";")

	// 4️⃣ 参数
	args = append(args,
		"-filter_complex", filter,
		"-c:a", "aac",
		"-b:a", "192k",
		output,
	)

	_, _, err := ffmpeg.RunFfmpeg(true, args...)
	return err
}

// 调整音量
func volume(input, output string, volume float64) error {
	if volume <= 0 {
		return nil
	}
	if output == "" {
		imputtmp := filepath.Join(filepath.Dir(input), strings.ReplaceAll(filepath.Base(input), ".", "--temp"))
		if err := os.Rename(input, imputtmp); err != nil {
			return err
		}
		output = input
		input = imputtmp
	}

	filter := fmt.Sprintf("volume=%.2f,alimiter=limit=0.95", volume)
	_, _, err := ffmpeg.RunFfmpeg(true,
		"-i", input,
		"-filter:a", filter,
		"-c:a", "aac",
		"-b:a", "192k",
		output,
	)
	return err
}
