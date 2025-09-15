package funcs

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	// "syscall"
	// "github.com/lxn/win"
	// "github.com/kbinani/screenshot"
)

// 获取运行的根路径
func RunnerPath() string {
	execFile, _ := os.Executable()
	// if err != nil {
	// 	return "", err
	// }
	return filepath.Dir(execFile)
}

func HiddenDir(path string) error {
	if runtime.GOOS == "windows" {
		Hide(path)
	}
	return nil
}

// FreePort 获取随机端口
func FreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return port, nil
}

// 标准 Base64 解编码
func Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// Base64Decode 标准 Base64 解码
func Base64Decode(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RunCommand 启动指定可执行文件，并传入任意参数
type Command struct {
	Dir     string        `json:"dir"`     // 执行目录
	Name    string        `json:"name"`    // 执行的语句
	Args    []string      `json:"args"`    // 参数
	Env     []string      `json:"env"`     // 环境变量（可选）
	Stdout  io.Writer     `json:"-"`       // 输出
	Stderr  io.Writer     `json:"-"`       // 错误输出
	Stdin   io.Reader     `json:"-"`       // 标准输入
	Timeout time.Duration `json:"timeout"` // 可选：超时时间
}

func (c *Command) Run() error {
	ctx := context.Background()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), c.Env...)

	if c.Stdout != nil {
		cmd.Stdout = c.Stdout
	}
	if c.Stderr != nil {
		cmd.Stderr = c.Stderr
	}
	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	}

	return cmd.Run()
}

func RunCommand(wait bool, cmdName string, args ...string) (string, error) {
	cmd := exec.Command(cmdName, args...)

	if wait {
		// 等待执行完成并获取输出
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return stderr.String(), fmt.Errorf("执行失败: %v", err)
		}
		return out.String(), nil
	} else {
		// 异步执行，不等待结果
		err := cmd.Start()
		if err != nil {
			return "", fmt.Errorf("启动失败: %v", err)
		}
		// 不等待，立即返回
		return "", nil
	}
}

// 解压 ZIP 文件到指定目录
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// 如果是目录，创建目录
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
			continue
		}

		// 创建上级目录
		os.MkdirAll(filepath.Dir(fpath), 0755)

		// 创建文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取当前的系统
func SysType() string {
	if runtime.GOOS == "windows" {
		return "windows"
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			return "macm"
		} else {
			return "mac"
		}
	} else if runtime.GOOS == "linux" {
		return "linux"
	} else {
		return runtime.GOOS
	}
}

// 获取本机中局域网的ip
func GetLocalIP() (string, error) {
	// 获取所有网卡的地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// 断言为 *net.IPNet 类型
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		// 只要 IPv4，并且不是回环地址
		if ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
			ip := ipNet.IP.String()
			if strings.Contains(ip, "192.168") {
				return ip, nil
			}
		}
	}
	return "", fmt.Errorf("未找到局域网 IP")
}
