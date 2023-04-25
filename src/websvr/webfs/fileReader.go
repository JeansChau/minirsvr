package webfs

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func LoadFile(path string) []string {
	// 打开指定文件夹
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModeDir)
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(0)
	}
	defer f.Close()
	// read all files in the dir
	fileInfo, _ := f.ReadDir(-1)

	files := make([]string, 0)
	for _, info := range fileInfo {
		files = append(files, info.Name())
	}
	return files
}

func FileReader(pathStr *string) *SendBufInfo {

	sendBuf := emptyBuf

	success := true
	var fiPtr *fs.FileInfo = nil
	// fmt.Println("FileReader(), pathStr", *pathStr)
	buf, err := ioutil.ReadFile((*pathStr))
	if err == nil {
		fi, err := os.Stat(*pathStr)
		if err == nil {
			fiPtr = &fi
			sendBuf = buf
		}

	} else {
		fmt.Println("Error: ", err)
		success = false
	}
	contentType := http.DetectContentType(buf)
	// gzipEnabled := true
	// fmt.Println("FileReader(), have not gzip")
	gzipEnabled := false
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "application/octet-stream", "image/x-icon":
		gzipEnabled = false
	default:
	}
	si := SendBufInfo{&sendBuf, len(buf), fiPtr, contentType, gzipEnabled, success}
	return &si
}
func GzipSendBuf(siPtr *SendBufInfo) {

	if siPtr.GzipEnabled && siPtr.Success {
		var zBuf bytes.Buffer
		zw := gzip.NewWriter(&zBuf)

		if _, err := zw.Write(*siPtr.BufPtr); err != nil {
			siPtr.GzipEnabled = false
			fmt.Println("gzip is faild,err:", err)
			zw.Close()
		} else {
			zw.Close()
			zb := zBuf.Bytes()
			siPtr.BufPtr = &zb
			siPtr.Size = len(zb)
			// fmt.Println("fileResponse(), does gzip, si.Size: ", siPtr.Size)

		}
	}
}
