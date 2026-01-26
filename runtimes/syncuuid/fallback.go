package syncuuid

import (
	"net"
	"os"
)

func fallbackID() string {
	if mac := firstMAC(); mac != "" {
		return mac
	}
	if hn, err := os.Hostname(); err == nil {
		return hn
	}
	return ""
}

func firstMAC() string {
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if i.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(i.HardwareAddr) > 0 {
			return i.HardwareAddr.String()
		}
	}
	return ""
}
