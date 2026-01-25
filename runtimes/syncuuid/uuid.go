package syncuuid

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sys/windows/registry"
)

// MachineUUID 返回稳定的机器 UUID
func MachineUUID() string {
	// 1. 尝试系统机器 ID
	if id := getMachineID(); id != "" {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(id)).String()
	}
	// 2. 尝试 MAC 地址
	if id := getMACBasedID(); id != "" {
		return uuid.NewMD5(uuid.Nil, []byte(id)).String()
	}
	// 3. 最后备用随机
	return uuid.New().String()
}

// getMachineID 从操作系统取机器唯一 ID
func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		return readLinuxMachineID()
	case "windows":
		return readWindowsMachineID()
	case "darwin":
		return readDarwinMachineID()
	}
	return ""
}

func readLinuxMachineID() string {
	paths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			if id := strings.TrimSpace(string(data)); id != "" {
				return id
			}
		}
	}
	return ""
}

func readWindowsMachineID() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()
	s, _, _ := key.GetStringValue("MachineGuid")
	return strings.TrimSpace(s)
}

func readDarwinMachineID() string {
	// macOS 使用 ioreg 获取 IOPlatformUUID
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// 解析 UUID
	for _, line := range bytes.Split(out, []byte("\n")) {
		if bytes.Contains(line, []byte("IOPlatformUUID")) {
			parts := bytes.Split(line, []byte("\""))
			if len(parts) >= 4 {
				return string(parts[3])
			}
		}
	}
	return ""
}

// getMACBasedID 备选方案：取第一个 MAC
func getMACBasedID() string {
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if len(i.HardwareAddr) == 0 {
			continue
		}
		return i.HardwareAddr.String()
	}
	return ""
}
