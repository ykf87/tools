//go:build windows

package hideconsole

import "syscall"

func HideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := proc.Call()
	if hwnd != 0 {
		user32 := syscall.NewLazyDLL("user32.dll")
		proc := user32.NewProc("ShowWindow")
		proc.Call(hwnd, 0)
	}
}
