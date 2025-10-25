package file

import (
	"net/http"
	"os"
	"strings"
	"time"
	"tools/runtimes/funcs"
)

type File struct {
	FullName string
	reader   *os.File
	info     os.FileInfo
}

func NewFileInfo(f string) (*File, error) {
	fh, err := os.Open(f)
	if err != nil {
		return nil, err
	}

	fl := new(File)
	fl.FullName = f
	fl.reader = fh
	fl.info, _ = os.Stat(f)
	return fl, nil
}

func (this *File) Close() {
	if this.reader != nil {
		this.reader.Close()
	}
}

func (this *File) GetMime() string {
	buf := make([]byte, 512)
	if _, err := this.reader.Read(buf); err == nil {
		return strings.ToLower(http.DetectContentType(buf))
	}
	return ""
}

func (this *File) Size() int64 {
	return this.info.Size()
}

func (this *File) Time() time.Time {
	return this.info.ModTime()
}

func (this *File) Md5() string {
	return funcs.Md5File(this.reader)
}
