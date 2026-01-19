package minio

import (
	"fmt"
	"os"
	"os/exec"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
)

var MINIPORT int

func init() {
	MINIPORT, _ = funcs.FreePort()
	funcs.RunCommandWithENV(false, config.FullPath(config.SYSROOT, "minio.exe"), func(cmd *exec.Cmd) {
		cmd.Env = append(os.Environ(),
			"MINIO_ROOT_USER=admin",
			"MINIO_ROOT_PASSWORD=StrongPassword123!",
		)
	}, "server", config.FullPath(".mini"), "--console-address", fmt.Sprintf("0.0.0.0:%d", MINIPORT))
	fmt.Println("minio启动端口:", MINIPORT)
}
