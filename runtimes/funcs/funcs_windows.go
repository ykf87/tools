//go:build windows

package funcs

import (
	"os"
	"syscall"
)

func Hide(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		//Windows 下设置隐藏属性
		p, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return err
		}
		// 设置隐藏属性
		err = syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_HIDDEN)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}
