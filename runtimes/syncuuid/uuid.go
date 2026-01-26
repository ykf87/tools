package syncuuid

import (
	"crypto/sha256"
	"encoding/hex"
)

func MachineUUID() string {
	raw := getRawMachineID()
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// MachineUUID 返回稳定的机器 UUID
// func MachineUUID() string {
// 	// 1. 尝试系统机器 ID
// 	if id := getMachineID(); id != "" {
// 		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(id)).String()
// 	}
// 	// 2. 尝试 MAC 地址
// 	if id := getMACBasedID(); id != "" {
// 		return uuid.NewMD5(uuid.Nil, []byte(id)).String()
// 	}
// 	// 3. 最后备用随机
// 	return uuid.New().String()
// }

// // getMachineID 从操作系统取机器唯一 ID
// func getMachineID() string {
// 	switch runtime.GOOS {
// 	case "linux":
// 		return readLinuxMachineID()
// 	case "windows":
// 		return readWindowsMachineID()
// 	case "darwin":
// 		return readDarwinMachineID()
// 	}
// 	return ""
// }

// func readLinuxMachineID() string {
// 	paths := []string{
// 		"/etc/machine-id",
// 		"/var/lib/dbus/machine-id",
// 	}
// 	for _, p := range paths {
// 		if data, err := os.ReadFile(p); err == nil {
// 			if id := strings.TrimSpace(string(data)); id != "" {
// 				return id
// 			}
// 		}
// 	}
// 	return ""
// }
