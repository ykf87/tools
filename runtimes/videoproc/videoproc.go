package videoproc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/downloader"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/mainsignal"
)

type VideoConfig struct {
	Input  string
	Output string

	CRF    int
	Preset string

	Audio string
}

type ProbeResult struct {
	Streams []struct {
		RFrameRate string `json:"r_frame_rate"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
	} `json:"streams"`
}

func ProbeVideo(path string) (*ProbeResult, error) {
	out, _, err := ffmpeg.RunFfporbe(
		true,
		"-v",
		"error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate",
		"-of", "json",
		path,
	)
	if err != nil {
		return nil, err
	}

	var res ProbeResult
	err = json.Unmarshal([]byte(out), &res)
	return &res, err
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func BuildFilter(width, height int) string {

	brightness := randFloat(-0.03, 0.03)
	contrast := randFloat(0.95, 1.08)
	saturation := randFloat(0.95, 1.15)

	noise := randInt(5, 12)

	cropX := randInt(0, 6)
	cropY := randInt(0, 6)

	speed := randFloat(0.97, 1.03)

	textX := randInt(0, width/2)
	textY := randInt(0, height/2)

	return fmt.Sprintf(`
setpts=%f*PTS,
eq=brightness=%f:contrast=%f:saturation=%f,
noise=alls=%d:allf=t,
crop=iw-%d:ih-%d,
scale=iw:ih,
drawtext=text='%d':x=%d:y=%d:fontsize=24:fontcolor=white@0.15
`,
		speed,
		brightness, contrast, saturation,
		noise,
		cropX, cropY,
		randInt(1000, 9999),
		textX, textY,
	)
}

func BuildFilter1(width, height int) string {

	brightness := randFloat(-0.02, 0.02)
	contrast := randFloat(0.98, 1.05)
	saturation := randFloat(0.98, 1.08)

	speed := randFloat(0.98, 1.02)

	// ✅ 根据分辨率限制裁剪幅度（关键）
	maxCropX := width / 100 // 最多裁 1%
	maxCropY := height / 100

	if maxCropX < 2 {
		maxCropX = 2
	}
	if maxCropY < 2 {
		maxCropY = 2
	}

	var sc string
	if width > height {
		sc = "1280:720"
	} else {
		sc = "720:1280"
	}

	// ✅ 强制偶数
	cropX := randInt(0, maxCropX/2) * 2
	cropY := randInt(0, maxCropY/2) * 2

	return fmt.Sprintf(
		"setpts=%f*PTS,eq=brightness=%f:contrast=%f:saturation=%f,crop=iw-%d:ih-%d,scale=%s,unsharp=5:5:0.5:5:5:0.0,setsar=1",
		speed,
		brightness, contrast, saturation,
		cropX, cropY,
		sc,
	)
}

func ProcessVideo(cfg VideoConfig) error {
	if cfg.Audio != "" {
		if strings.HasPrefix(cfg.Audio, "http") {
			fp := config.FullPath(config.MEDIAROOT, ".tmp")
			df, err := downloader.Download(mainsignal.MainCtx, &downloader.DownloadOption{
				URL: cfg.Audio,
				Dir: fp,
			})
			if err != nil {
				return err
			}
			cfg.Audio = df.FullName
			defer func() {
				os.Remove(cfg.Audio)
			}()
		} else {
			cfg.Audio = ""
		}
	}

	probe, err := ProbeVideo(cfg.Input)
	if err != nil {
		return err
	}

	stream := probe.Streams[0]
	filter := BuildFilter1(stream.Width, stream.Height)

	// 随机帧率（更像真实设备）
	fps := "30"
	if rand.Intn(2) == 0 {
		fps = "29.97"
	}

	// creation_time 随机（最近7天）
	ct := time.Now().Add(-time.Duration(randInt(0, 7)) * 24 * time.Hour).UTC().Format(time.RFC3339)

	// var args []string
	args := []string{
		"-fflags", "+genpts",
	}

	if cfg.Audio != "" {
		args = append(args,
			"-stream_loop", "-1",
			"-i", cfg.Audio,
		)
	}

	args = append(args,
		"-i", cfg.Input,
		"-filter:v", filter,
	)

	// map（必须完整）
	if cfg.Audio != "" {
		args = append(args, "-map", "1:v", "-map", "0:a")
	} else {
		args = append(args, "-map", "0:v", "-map", "0:a?")
	}

	// metadata
	args = append(args,
		"-map_metadata", "-1",
		"-map_chapters", "-1",
		"-movflags", "+faststart+use_metadata_tags",

		"-metadata", "major_brand=mp42",
		"-metadata", "minor_version=1",
		"-metadata", "compatible_brands=isommp41mp42",
		"-metadata", "creation_time="+ct,
		"-metadata", "encoder=Apple QuickTime",

		"-metadata:s:v:0", "handler_name=Core Media Video",
		"-metadata:s:a:0", "handler_name=Core Media Audio",
		"-metadata:s:v:0", "encoder=",
	)

	// video
	if cfg.Preset == "" {
		cfg.Preset = "slow"
	}
	if cfg.CRF == 0 {
		cfg.CRF = 20
	}
	args = append(args,
		"-c:v", "libx264",
		"-tag:v", "avc1",
		"-g", "60",
		"-keyint_min", "30",
		"-sc_threshold", "40",
		"-crf", fmt.Sprintf("%d", cfg.CRF),
		"-maxrate", "2800k",
		"-bufsize", "5600k",
		"-preset", cfg.Preset, //slow
		"-r", fps,
	)

	// audio（只在有音频时处理）
	if cfg.Audio != "" {
		args = append(args,
			"-c:a", "aac",
			"-b:a", "128k",
			"-ac", "2",
			"-ar", "44100",
			"-af", "volume=1.01",
			"-shortest",
		)
	}

	// 收尾
	args = append(args,
		"-avoid_negative_ts", "make_zero",
		cfg.Output,
	)

	_, _, err = ffmpeg.RunFfmpeg(true, args...)
	// _, _, err = ffmpeg.RunFfmpeg(
	// 	true,
	// 	"-fflags", "+genpts",
	// 	"-i", cfg.Input,
	// 	"-vf", filter,

	// 	// ===== metadata 清洗 =====
	// 	"-map_metadata", "-1",
	// 	"-map_chapters", "-1",
	// 	"-movflags", "+faststart+use_metadata_tags",

	// 	// ===== metadata 伪装 =====
	// 	"-metadata", "major_brand=mp42",
	// 	"-metadata", "minor_version=1",
	// 	"-metadata", "compatible_brands=isommp41mp42",
	// 	"-metadata", "creation_time="+ct,
	// 	"-metadata", "encoder=Apple QuickTime",

	// 	"-metadata:s:v:0", "handler_name=Core Media Video",
	// 	"-metadata:s:a:0", "handler_name=Core Media Audio",
	// 	"-metadata:s:v:0", "encoder=",

	// 	// ===== 视频编码 =====
	// 	"-c:v", "libx264",
	// 	"-tag:v", "avc1",
	// 	// "-preset", "medium",

	// 	// GOP结构（关键）
	// 	"-g", "60",
	// 	"-keyint_min", "30",
	// 	"-sc_threshold", "40",

	// 	// 码率控制（稳定且真实）
	// 	// "-b:v", "1200k",
	// 	// "-maxrate", "1500k",
	// 	// "-bufsize", "3000k",
	// 	"-crf", "20",
	// 	"-maxrate", "2800k",
	// 	"-bufsize", "5600k",
	// 	"-preset", "slow",

	// 	// 帧率
	// 	"-r", fps,

	// 	// ===== 音频 =====
	// 	"-c:a", "aac",
	// 	"-af", "volume=1.01",

	// 	// 时间戳处理
	// 	"-avoid_negative_ts", "make_zero",

	// 	cfg.Output,
	// )
	return err
}
