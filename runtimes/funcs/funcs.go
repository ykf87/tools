package funcs

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	// "syscall"
	// "github.com/lxn/win"
	// "github.com/kbinani/screenshot"
	"github.com/dop251/goja"
	"github.com/google/uuid"
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

// 检查端口是否被占用
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false // 已被占用
	}
	_ = ln.Close()
	return true
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

// urldecode 解码
func UrlDecode(str string) string {
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		return ""
	}
	return decoded
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

func RunCommand(wait bool, cmdName string, args ...string) (string, *exec.Cmd, error) {
	cmd := exec.Command(cmdName, args...)

	if wait {
		// 等待执行完成并获取输出
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return stderr.String(), nil, fmt.Errorf("执行失败: %v", err)
		}
		return out.String(), cmd, nil
	} else {
		// 异步执行，不等待结果
		err := cmd.Start()
		if err != nil {
			return "", nil, fmt.Errorf("启动失败: %v", err)
		}
		// 不等待，立即返回
		return "", cmd, nil
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
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "macm"
		} else {
			return "mac"
		}
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
	// if runtime.GOOS == "windows" {
	// 	return "windows"
	// } else if runtime.GOOS == "darwin" {
	// 	if runtime.GOARCH == "arm64" {
	// 		return "macm"
	// 	} else {
	// 		return "mac"
	// 	}
	// } else if runtime.GOOS == "linux" {
	// 	return "linux"
	// } else {
	// 	return runtime.GOOS
	// }
}

// 获取本机中局域网的ip
func GetLocalIP(local bool) (string, error) {
	if local == false {
		return "127.0.0.1", nil
	}
	// return "127.0.0.1", nil
	// 获取所有网卡的地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1", err
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
			ips := strings.Split(ip, ".")
			if ips[len(ips)-1] == "1" {
				continue
			}
			// fmt.Println(ip, "--------")
			if strings.Contains(ip, "192.168") {
				return ip, nil
			}
		}
	}
	return "127.0.0.1", nil
}

// 获取当前固定的uuid
func Uuid() string {
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		if len(i.HardwareAddr) == 0 {
			continue
		}
		// 用 MAC 地址生成固定 UUID
		return uuid.NewMD5(uuid.Nil, i.HardwareAddr).String()
	}
	return uuid.New().String()
}

// 生产随机的uuid
func RoundmUuid() string {
	uid, err := uuid.NewUUID()
	if err != nil {
		return ""
		// panic(err)
	}
	return uid.String()
}

// 生成密码
func GenPassword(password string, cost int) (string, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(hash), err
}

// 验证密码
func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// 截断保留decimals位小数
func TruncateFloat64(f float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Trunc(f*factor) / factor
}

// 字符串md5
func Md5String(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))

	return hex.EncodeToString(hash.Sum(nil))
}

// 文件md5
func Md5File(reader io.Reader) string {
	hash := md5.New()
	const bufferSize = 8192 // 8 KB 缓冲区
	buffer := make([]byte, bufferSize)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("File Read Failure:" + err.Error())
			return ""
		}
		if n == 0 {
			break // 数据读取完成
		}
		hash.Write(buffer[:n]) // 写入哈希对象
	}

	// 返回计算出的 MD5 值
	return hex.EncodeToString(hash.Sum(nil))
}

// RandomMAC 生成一个随机 MAC 地址
// prefix 可以为空，例如 "00:1A:2B" 表示指定厂商前缀
func RandomMAC(prefix string) string {
	rand.Seed(time.Now().UnixNano())

	mac := make([]byte, 6)

	if prefix != "" {
		parts := strings.Split(prefix, ":")
		if len(parts) != 3 && len(parts) != 6 {
			panic("prefix 格式错误，例如：00:1A:2B 或 00:1A:2B:XX:XX:XX")
		}
		for i := 0; i < len(parts) && i < 6; i++ {
			fmt.Sscanf(parts[i], "%02X", &mac[i])
		}
		rand.Read(mac[len(parts):])
	} else {
		rand.Read(mac)
		// 设置本地管理位 & 单播位
		mac[0] = (mac[0] | 2) & 0xfe
	}

	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// 生成随机字符串
func NewNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// 打开目录
func OpenDir(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path) // Linux 下通常使用 xdg-open
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

const (
	_          = iota
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

func FormatFileSize(size int64) string {

	floatSize := float64(size)

	switch {
	case floatSize >= TB:
		return fmt.Sprintf("%.2f T", floatSize/TB)
	case floatSize >= GB:
		return fmt.Sprintf("%.2f G", floatSize/GB)
	case floatSize >= MB:
		return fmt.Sprintf("%.2f M", floatSize/MB)
	case floatSize >= KB:
		return fmt.Sprintf("%.2f K", floatSize/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// 正则替换内容
func ReplaceContent(content, prefix, suffix, k string, value any) string {
	// 对前后缀进行正则安全转义
	p := regexp.QuoteMeta(prefix)
	s := regexp.QuoteMeta(suffix)

	// 匹配：前缀 + 任意空格 + 任意内容 + 任意空格 + 后缀
	pattern := p + `\s*` + k + `\s*` + s
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllString(content, fmt.Sprint(value))
}

// 执行js并返回结果
func RunJs(js string) (any, error) {
	vm := goja.New()

	result, err := vm.RunString(`
			function add(a, b) {
				return a + b;
			}
			add(2, 3);
		`)

	if err != nil {
		return nil, err
	}
	return result.Export(), nil
}
