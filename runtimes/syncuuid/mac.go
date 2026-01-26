//go:build darwin
// +build darwin

package syncuuid

import (
	"bytes"
	"net"
	"os/exec"
	"strings"
)

func getRawMachineID() string {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	out, err := cmd.Output()
	if err == nil {
		if id := parseIORegUUID(out); id != "" {
			return id
		}
	}
	return fallbackID()
}

func parseIORegUUID(out []byte) string {
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				return strings.TrimSpace(string(parts[3]))
			}
		}
	}
	return ""
}

func readWindowsMachineID() string {
	return ""
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
