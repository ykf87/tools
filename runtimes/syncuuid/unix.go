//go:build linux || freebsd || openbsd || netbsd

package syncuuid

import (
	"os"
	"strings"
)

func getRawMachineID() string {
	paths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}
	for _, p := range paths {
		if b, err := os.ReadFile(p); err == nil {
			id := strings.TrimSpace(string(b))
			if id != "" {
				return id
			}
		}
	}
	return fallbackID()
}
