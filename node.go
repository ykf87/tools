package main

//
// import (
// 	"archive/tar"
// 	"archive/zip"
// 	// "bytes"
// 	"compress/gzip"
// 	// "fmt"
// 	"io"
// 	// "net/http"
// 	"os"
// 	"path/filepath"
// 	// "runtime"
// 	// "strings"
// 	"tools/runtimes/config"
// 	// "tools/runtimes/funcs"
// 	// "tools/runtimes/logs"
// )
//
// const nodeVersion = "20.11.1"
//
// var webFullPath string
// var nodePath string
//
// func init() {
// 	webFullPath = config.FullPath(config.WEBROOT)
// 	nodePath = filepath.Join(webFullPath, "nodejs")
// }
//
// // // 检查系统是否安装了nodejs
// // func checkNode() {
// // 	if _, err := os.Stat(nodePath); err != nil {
// // 		if err := installNode(nodeVersion, nodePath); err != nil {
// // 			logs.Error(err.Error())
// // 			panic(err)
// // 		}
// // 	}
// //
// // 	if err := npmInstall(); err != nil {
// // 		logs.Error(err.Error())
// // 		panic(err)
// // 	}
// //
// // 	// var stdout bytes.Buffer
// // 	// multiOut := io.MultiWriter(&stdout)
// //
// // 	// cmd := &funcs.Command{
// // 	// 	Dir:    webFullPath,
// // 	// 	Name:   "node",
// // 	// 	Args:   []string{"-v"},
// // 	// 	Stdout: multiOut,
// // 	// 	Stderr: multiOut,
// // 	// }
// // 	// if err := cmd.Run(); err != nil { //nodejs不存在,安装
// // 	// 	if err := installNode(); err != nil {
// // 	// 		logs.Error(err.Error())
// // 	// 		panic("nodejs install error!")
// // 	// 	}
// // 	// }
// // 	// fmt.Println(stdout.String())
// //
// // }
//
// // // 安装nodejs
// // func installNode(version, installDir string) error {
// // 	url, archiveName := getDownloadURL(version)
// // 	fmt.Println("下载 Node.js:", url)
// //
// // 	// 下载文件
// // 	resp, err := http.Get(url)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	defer resp.Body.Close()
// //
// // 	tmpFile := filepath.Join(os.TempDir(), archiveName)
// // 	out, err := os.Create(tmpFile)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	_, err = io.Copy(out, resp.Body)
// // 	out.Close()
// // 	if err != nil {
// // 		return err
// // 	}
// //
// // 	// 解压
// // 	if strings.HasSuffix(tmpFile, ".zip") {
// // 		return unzip(tmpFile, installDir)
// // 	} else if strings.HasSuffix(tmpFile, ".tar.gz") {
// // 		return untarGz(tmpFile, installDir)
// // 	}
// // 	return fmt.Errorf("未知的文件格式: %s", tmpFile)
// // }
//
// // // 检查依赖是否安装
// // func npmInstall() error {
// // 	packagejson := filepath.Join(webFullPath, "package.json")
// // 	if _, err := os.Stat(packagejson); os.IsNotExist(err) {
// // 		fmt.Println("下载组件...")
// // 		//config.WebFilesDownUrl
// // 		return nil
// // 	}
// // 	nodeModules := filepath.Join(webFullPath, "node_modules") //filepath.Join(projectPath, "node_modules")
// // 	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
// // 		fmt.Println("检测到 node_modules 不存在，正在执行 npm install...")
// //
// // 		var stdout bytes.Buffer
// // 		multiOut := io.MultiWriter(os.Stdout, &stdout)
// //
// // 		cmd := &funcs.Command{
// // 			Dir:    webFullPath,
// // 			Name:   "npm",
// // 			Args:   []string{"install"},
// // 			Stdout: multiOut,
// // 			Stderr: multiOut,
// // 		}
// // 		if err := cmd.Run(); err != nil {
// // 			logs.Error(err.Error())
// // 			return err
// // 		}
// // 	}
// // 	return nil
// // }
// //
// // // 获取nodejs下载地址
// // func getDownloadURL(version string) (url string, fileName string) {
// // 	ver := "v" + version
// // 	var osName, arch string
// //
// // 	switch runtime.GOOS {
// // 	case "windows":
// // 		osName = "win"
// // 	case "darwin":
// // 		osName = "darwin"
// // 	case "linux":
// // 		osName = "linux"
// // 	}
// //
// // 	switch runtime.GOARCH {
// // 	case "amd64":
// // 		arch = "x64"
// // 	case "arm64":
// // 		arch = "arm64"
// // 	default:
// // 		arch = runtime.GOARCH
// // 	}
// //
// // 	if osName == "win" {
// // 		fileName = fmt.Sprintf("node-%s-%s-%s.zip", ver, osName, arch)
// // 	} else {
// // 		fileName = fmt.Sprintf("node-%s-%s-%s.tar.gz", ver, osName, arch)
// // 	}
// //
// // 	url = fmt.Sprintf("https://nodejs.org/dist/%s/%s", ver, fileName)
// // 	return
// // }
//
// // 解压 zip
// func unzip(src, dest string) error {
// 	r, err := zip.OpenReader(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer r.Close()
//
// 	for _, f := range r.File {
// 		fpath := filepath.Join(dest, f.Name)
// 		if f.FileInfo().IsDir() {
// 			os.MkdirAll(fpath, os.ModePerm)
// 			continue
// 		}
// 		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
// 			return err
// 		}
//
// 		rc, err := f.Open()
// 		if err != nil {
// 			return err
// 		}
// 		outFile, err := os.Create(fpath)
// 		if err != nil {
// 			return err
// 		}
// 		_, err = io.Copy(outFile, rc)
//
// 		outFile.Close()
// 		rc.Close()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
//
// // 解压 tar.gz
// func untarGz(src, dest string) error {
// 	f, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
//
// 	gzr, err := gzip.NewReader(f)
// 	if err != nil {
// 		return err
// 	}
// 	defer gzr.Close()
//
// 	tr := tar.NewReader(gzr)
// 	for {
// 		header, err := tr.Next()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return err
// 		}
//
// 		target := filepath.Join(dest, header.Name)
// 		switch header.Typeflag {
// 		case tar.TypeDir:
// 			os.MkdirAll(target, os.FileMode(header.Mode))
// 		case tar.TypeReg:
// 			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
// 				return err
// 			}
// 			outFile, err := os.Create(target)
// 			if err != nil {
// 				return err
// 			}
// 			_, err = io.Copy(outFile, tr)
// 			outFile.Close()
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
