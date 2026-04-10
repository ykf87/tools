package minio

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/services"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MAINFILENAME = "minio.exe"
	PACKNAME     = "minio.zip"
)

var runcmd *exec.Cmd
var LanIp string

func init() {
	fn := fullName()
	if _, err := os.Stat(fn); err != nil {
		if err := download(fn); err != nil {
			panic(err)
		}
	}
	LanIp, _ = funcs.GetLocalIP(true)

	start(fn)
}

func Flush() {
	if runcmd != nil {
		runcmd.Process.Kill()
	}
}

func ControlUrl(tp string) string {
	ipaddr := "0.0.0.0"
	switch tp {
	case "nat":
		ipaddr = LanIp
	case "local":
		ipaddr = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", ipaddr, config.MINIPORT)
}

func ApiUrl(tp string) string {
	ipaddr := "0.0.0.0"
	switch tp {
	case "nat":
		ipaddr = LanIp
	case "local":
		ipaddr = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", ipaddr, config.MINIAPIPORT)
}

func start(fn string) {
	_, cmd, err := funcs.RunCommandWithENV(false,
		fn,
		func(cmd *exec.Cmd) {
			cmd.Env = append(os.Environ(),
				"MINIO_ROOT_USER="+config.ACCESSKEY,
				"MINIO_ROOT_PASSWORD="+config.SECRETKEY,
			)
		}, "server",
		config.FullPath(config.MINISAVE),
		"--console-address", ControlUrl(""),
		"--address", ApiUrl(""),
	)
	if err != nil {
		panic(err)
	}
	runcmd = cmd

	if err := createBucket(config.BUCKET); err == nil {
		// fmt.Println("io文件存储启动端口:", config.MINIPORT)
		config.DefStorage = "minio"
	} else {
		config.DefStorage = "local"
	}
}

func fullName() string {
	return config.FullPath(config.SYSROOT, MAINFILENAME)
}

func download(fn string) error {
	fmt.Println("下载 文件存储系统...")
	if err := services.ServerDownload(PACKNAME, filepath.Dir(fn), nil, func(total, downloaded, speed, workers int64) {
		msgstr := fmt.Sprintf(
			"%.2f%% %s/s %s 线程: %d",
			float64(downloaded)/float64(total)*100,
			funcs.FormatFileSize(speed, "1", ""),
			funcs.FormatFileSize(total, "1", ""),
			workers,
		)
		fmt.Print("\r", msgstr)
	}); err != nil {
		return err
	}
	fmt.Println("\n文件存储系统 下载成功! 解压...")

	pkn := config.FullPath(config.SYSROOT, PACKNAME)
	if err := funcs.Unzip(pkn, filepath.Dir(fn)); err != nil {
		fmt.Println("解压失败:", err)
		return err
	}
	return os.Remove(pkn)
}

func waitPort(addr string) {
	startTime := time.Now().Unix()
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(500 * time.Millisecond)
		if time.Now().Unix()-startTime >= 30 {
			break
		}
	}
	panic("minio not ready")
}

// 创建桶
func createBucket(bucketName string) error {
	endpoint := fmt.Sprintf("0.0.0.0:%d", config.MINIAPIPORT)
	waitPort(endpoint)
	// 创建客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.ACCESSKEY, config.SECRETKEY, ""),
		Secure: config.USESSL,
	})
	if err != nil {
		return err
	}

	location := "us-east-1"

	// 检查 bucket 是否存在
	exists, err := client.BucketExists(mainsignal.MainCtx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		// 创建 bucket
		err = client.MakeBucket(mainsignal.MainCtx, bucketName, minio.MakeBucketOptions{
			Region: location,
		})
		if err != nil {
			return err
		}

		// 设置为公共
		policy := `{
 "Version":"2012-10-17",
 "Statement":[
  {
   "Effect":"Allow",
   "Principal":"*",
   "Action":["s3:GetObject"],
   "Resource":["arn:aws:s3:::` + bucketName + `/*"]
  }
 ]
}`
		err = client.SetBucketPolicy(mainsignal.MainCtx, bucketName, policy)
		if err != nil {
			return err
		}
	}
	return nil
}
