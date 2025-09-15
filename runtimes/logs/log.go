package logs

import (
	"os"
	"path/filepath"
	"time"
	"tools/runtimes/config"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const LOGPATH = "logs/"

var Logger zerolog.Logger

func init() {
	lgp := config.FullPath(config.LOGROOT)
	if err := os.MkdirAll(lgp, os.ModePerm); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to create log directory")
	}

	// 创建按天切分的日志文件
	logFileName := filepath.Join(lgp, "app.log")
	logWriter, err := rotatelogs.New(
		logFileName+".%Y%m%d",
		rotatelogs.WithLinkName(logFileName),
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 保留7天的日志
		rotatelogs.WithRotationTime(24*time.Hour), // 每天切分一次
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to create log file")
	}
	Logger = zerolog.New(logWriter).With().Timestamp().Logger()
}
func Info(msg string) {
	Logger.Info().Msg(msg)
}
func Error(msg string) {
	Logger.Error().Msg(msg)
}
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}
