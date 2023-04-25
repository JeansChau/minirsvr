package client

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"voxwebsvr.com/webfs"
)

// go mod init voxwebsvr.com/client
// go mod edit -replace voxwebsvr.com/webfs=../webfs

var unixEpochTime = time.Unix(0, 0)

func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}
func setLastModified(w http.ResponseWriter, modtime time.Time) {
	if !isZeroTime(modtime) {
		w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	}
}

var errorTemplate string = `
<!DOCTYPE html>
<html lang="en"><head></head>
<body><p align="center">Error: illegal request !!!</p></body>
`

func sendBytesToClient(w *http.ResponseWriter, siPtr *webfs.SendBufInfo) {

	wr := (*w)
	if siPtr.Success {

		if siPtr.FiPtr != nil {
			setLastModified(wr, (*siPtr.FiPtr).ModTime())
		}
		header := wr.Header()

		header.Set("Content-Type", siPtr.ContentType)
		header.Set("Server", "golang")
		header.Set("Accept-Ranges", "bytes")
		header.Set("Content-Length", strconv.Itoa(siPtr.Size))

		// log.Print("Content-Length", strconv.Itoa(siPtr.Size))

		if siPtr.GzipEnabled {
			header.Set("Accept-Encoding", "gzip,deflate")
			header.Set("Vary", "Accept-Encoding")
			header.Set("Content-encoding", "gzip")
		}

		wr.WriteHeader(http.StatusOK)
		wr.Write((*(siPtr.BufPtr)))

	} else {

		wr.WriteHeader(http.StatusOK)
		fmt.Fprintf(wr, errorTemplate)
	}
}
func rangeFileResponse(w *http.ResponseWriter, pathStr *string, bytesPosList []int) {

	// fmt.Println("rangeFileResponse(), pathStr", *pathStr)
	siPtr := webfs.ReadFileBySteps(pathStr, bytesPosList[0], bytesPosList[1])
	sendBytesToClient(w, siPtr)
}

func fileResponse(w *http.ResponseWriter, pathStr *string) {

	// fmt.Println("fileResponse(), pathStr", *pathStr)
	siPtr := webfs.FileReader(pathStr)
	webfs.GzipSendBuf(siPtr)
	// fmt.Println("fileResponse(),contentType: ", siPtr.ContentType, "siPtr.GzipEnabled: ", siPtr.GzipEnabled)
	sendBytesToClient(w, siPtr)

}
func ReceiveRequest(w *http.ResponseWriter, r *http.Request, pathStrPtr *string) {
	var rHeader = r.Header
	var rangeList []string
	rangeListSize := 0
	if _, v := rHeader["Range"]; v {
		rangeList = rHeader["Range"]
		rangeListSize = len(rangeList)
	}

	if rangeListSize > 0 {
		var bytesPosList []int
		for _, value := range rangeList {
			posList := strings.Split(value, "=")
			if posList[0] == "bytes" && len(posList) == 2 {
				posList := strings.Split(posList[1], "-")
				if len(posList) == 2 {
					num0, _ := strconv.Atoi(posList[0])
					num1, _ := strconv.Atoi(posList[1])
					bytesPosList = append(bytesPosList, num0, num1)
					break
				}
			}
		}
		rangeFileResponse(w, pathStrPtr, bytesPosList)
	} else {
		fileResponse(w, pathStrPtr)
	}
}
