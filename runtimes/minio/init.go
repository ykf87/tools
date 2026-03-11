package minio

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {

	funcs.RunCommandWithENV(false,
		config.FullPath(config.SYSROOT, "minio.exe"),
		func(cmd *exec.Cmd) {
			cmd.Env = append(os.Environ(),
				"MINIO_ROOT_USER="+config.ACCESSKEY,
				"MINIO_ROOT_PASSWORD="+config.SECRETKEY,
			)
		}, "server",
		config.FullPath(".mini"),
		"--console-address", fmt.Sprintf("0.0.0.0:%d", config.MINIPORT),
		"--address", fmt.Sprintf("0.0.0.0:%d", config.MINIAPIPORT),
	)

	if err := CreateBucket(config.BUCKET); err == nil {
		fmt.Println("io文件存储启动端口:", config.MINIPORT)
		config.DefStorage = "minio"
	} else {
		config.DefStorage = "local"
	}
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
func CreateBucket(bucketName string) error {
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
