package videoproc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tools/runtimes/clearer"
	"tools/runtimes/config"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/imager"
)

type limits struct {
	Queue chan byte
	// Done
}

var limit *limits

func init() {
	limit = new(limits)
	limit.Queue = make(chan byte, 1)
}

func SecMaker(videos []string, audio *AudioInpter) (*Maker, error) {
	mk := new(Maker)
	mk.Srcs = videos
	mk.Audio = audio
	mk.Factory = new(Factory)
	return mk, nil
}

func (m *Maker) Output(ctx context.Context, output string) error {
	m.ctx = ctx
	if err := m.buildTempDir(); err != nil {
		return err
	}

	if err := m.parseInfo(); err != nil {
		return err
	}

	defer func() {
		if strings.Contains(m.tempdir, ".tmp") {
			fmt.Println(os.RemoveAll(m.tempdir), "删除临时目录")
		}
	}()

	if err := m.splitVideoToImg(); err != nil {
		return err
	}
	if err := m.mkimgs(); err != nil {
		return err
	}

	if err := m.merge(output); err != nil {
		return err
	}

	return nil
}

func (m *Maker) merge(output string) error {
	fls, err := filepath.Glob(filepath.Join(m.framesDir, fmt.Sprintf("*%s", IMGTYPE)))
	if err != nil {
		return err
	}

	iip := filepath.Join(filepath.Dir(fls[0]), "merge_%06d."+IMGTYPE) //fmt.Sprintf("merge_%06d.%s", (k+1), IMGTYPE)
	for k, v := range fls {
		if err := os.Rename(v, fmt.Sprintf(iip, (k+1))); err != nil {
			return err
		}
	}

	if m.audio == nil || m.audio.Audio == nil {
		return errors.New("缺少音频文件")
	}

	if _, _, err := ffmpeg.RunFfmpeg(true,
		"-framerate", strconv.FormatFloat(m.srcs[0].Video.FPS, 'f', -1, 64),
		"-i", iip,
		"-i", m.audio.Audio.Url,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-shortest",
		output,
	); err != nil {
		return err
	}

	return nil
}

// (图片)帧处理
func (m *Maker) mkimgs() error {
	var task []Task
	var clearCleear []Task
	if m.Factory != nil {
		fls, err := filepath.Glob(filepath.Join(m.framesDir, fmt.Sprintf("*%s", IMGTYPE)))
		if err != nil {
			return err
		}
		for _, f := range fls {
			task = append(task, func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				img, _ := imager.NewImager(f)
				if m.Factory.Mirror == 1 || m.Factory.Mirror == 2 {
					img.Flip = &imager.Flip{
						Type: m.Factory.Mirror,
					}
				}
				if m.Factory.Crop != nil {
					img.Crop = m.Factory.Crop
				}
				if m.Factory.Linear != nil {
					img.Linear = m.Factory.Linear
				}
				if m.Factory.Resize != nil {
					img.Resize = m.Factory.Resize
				}
				if m.Factory.Rotation != nil {
					img.Rotation = m.Factory.Rotation
				}
				img.Width = m.Width
				img.Height = m.Height

				kwh := imager.KeepWH(true)
				img.KeepWH = &kwh
				return img.Output(f, 3)
			})

			if m.Factory.Clearer == true { // 变清晰和一般处理需要分开处理
				clearCleear = append(clearCleear, func(ctx context.Context) error {
					outname := f + ".png"
					if _, err := clearer.Clearers(f, outname, ""); err != nil {
						return err
					}
					defer func() {
						os.Remove(outname)
					}()

					return imager.RunBin("resize", outname, f, strconv.FormatFloat(0.25, 'f', -1, 64))
				})
			}
		}
	}

	if err := RunWithCancel(m.ctx, 4, task); err != nil {
		return err
	}
	if len(clearCleear) > 0 {
		RunWithCancel(m.ctx, 2, clearCleear)
	}
	return nil

	// fmt.Println(len(files), int(m.srcs[0].Video.FPS*m.srcs[0].Video.Duration))
}

// 将视频分解成图片
func (m *Maker) splitVideoToImg() error {
	for k, v := range m.srcs {
		if _, _, err := ffmpeg.RunFfmpeg(true,
			"-i", v.Video.Url,
			"-vsync", "0",
			"-q:v", "1",
			filepath.Join(m.framesDir, fmt.Sprintf("%d", k)+"_%06d."+IMGTYPE),
		); err != nil {
			return err
		}
		v.Video.ImgNumers = int(v.Video.FPS * v.Video.Duration)
	}
	return nil
}

// 构建临时处理目录
func (m *Maker) buildTempDir() error {
	m.tempdir = config.FullPath(config.MEDIAROOT, ".tmp", filepath.Base(m.Srcs[0]))
	m.framesDir = filepath.Join(m.tempdir, IMGDIR)
	if _, err := os.Stat(m.framesDir); err != nil {
		if err := os.MkdirAll(m.framesDir, os.ModePerm); err != nil {
			return nil
		}
	}
	return nil
}

// 先将传入的信息解析
func (m *Maker) parseInfo() error {
	for k, v := range m.Srcs {
		mf, err := probeMedia(v)
		if err != nil {
			return err
		}

		if m.Width <= 0 {
			m.Width = mf.Video.Width
		}
		if m.Height <= 0 {
			m.Height = mf.Video.Height
		}

		// 分离视频的音频
		VideoAudioSrc := filepath.Join(m.tempdir, fmt.Sprintf("%d.aac", k))
		_, _, err = ffmpeg.RunFfmpeg(true,
			"-fflags", "+genpts", // 强制生成时间戳
			"-i", v,
			"-vn",
			"-ac", "2",
			"-ar", "44100",
			"-c:a", "aac",
			"-b:a", "192k",
			"-af", "asetpts=PTS-STARTPTS",
			VideoAudioSrc,
		)
		if err == nil {
			if amf, err := probeMedia(VideoAudioSrc); err == nil {
				mf.Audio = amf.Audio
				// v.Audio = mf.Audio
			}
		} else {
			panic(err)
		}

		m.srcs = append(m.srcs, mf)
	}

	// 音频处理
	if m.Audio != nil && m.Audio.Url != "" { // 如果传入了替换音频,则将音频信息记录到最终音频
		ma := filepath.Join(m.tempdir, "audioinputer.aac")
		if m.Audio.Volume > 0 { // 如果传入的视频需要修改音量,则先调整音量
			if err := volume(m.Audio.Url, ma, m.Audio.Volume); err != nil {
				return err
			}
			m.Audio.Url = ma
			m.Audio.Volume = 1
		}

		mf, err := probeMedia(m.Audio.Url)
		if err != nil {
			return err
		}
		m.audio = mf
	} else { // 如果不要求替换音频,则将源视频的音频进行拼接并且稍微做一些干扰
		var auds []*AudioInpter
		for _, v := range m.srcs { // 获取视频的音频临时地址
			if v.Audio != nil {
				auds = append(auds, &AudioInpter{Url: v.Audio.Url})
			}
		}
		var aduioname string // 最终生成的一个音频文件(合并所有视频的音频后的文件)
		if len(auds) > 1 {
			aduioname = filepath.Join(m.tempdir, "merger.aac")
			if err := audioMerge(auds, aduioname); err != nil {
				return err
			}
		} else if len(auds) == 1 {
			aduioname = auds[0].Url
		} else {
			return errors.New("缺少音频文件!")
		}

		if aduioname != "" { // 对视频中音频做随机干扰
			outname := filepath.Join(m.tempdir, "audio.aac")
			if err := audioInterference(&AudioInpter{Url: aduioname}, outname); err != nil {
				return err
			}
			os.Remove(aduioname)
			amf, err := probeMedia(outname)
			if err != nil {
				return err
			}
			m.audio = amf
		}
	}

	// 混音处理(叠加音频)
	if m.AmixAudio != nil && m.AmixAudio.Url != "" {
		outname := filepath.Join(m.tempdir, "amxiAudio.aac")
		if err := audioAmix([]*AudioInpter{&AudioInpter{Url: m.audio.Audio.Url, Volume: 1.0}, m.AmixAudio}, outname); err == nil {
			if amf, err := probeMedia(outname); err == nil {
				os.Remove(m.audio.Audio.Url)
				m.audio = amf
			}
		}
	}

	// for _, v := range m.Audios {
	// 	mf, err := probeMedia(v)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	mf.Audio.Url = v
	// 	m.audios = append(m.audios, mf)
	// }
	return nil
}
