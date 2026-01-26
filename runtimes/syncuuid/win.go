//go:build windows
// +build windows

package syncuuid

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

func getRawMachineID() string {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return fallbackID()
	}
	defer key.Close()

	s, _, err := key.GetStringValue("MachineGuid")
	if err != nil {
		return fallbackID()
	}
	return strings.TrimSpace(s)
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
	return ""
}

// getMACBasedID 备选方案：取第一个 MAC
func getMACBasedID() string {
	return ""
}
