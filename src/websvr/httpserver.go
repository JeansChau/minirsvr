package main

import (
	"os"
	"runtime"

	// "path"

	"fmt"

	//"log"

	"net/http"

	"voxwebsvr.com/client"
)

// strconv.Itoa
// nohup ./httpserver &
// export PATH=$PATH:/usr/local/go/bin
// go build -o .\ ..\src\websvr\httpserver.go
// go build -o ../../bin/ httpserver.go
// go build ../src/websvr/httpserver.go
// go run ../src/websvr/httpserver.go
// go build -o ../../bin/ httpserver.go

// go mod init voxwebsvr.com/main
// go mod edit -replace voxwebsvr.com/webfs=./webfs
// go mod edit -replace voxwebsvr.com/client=./client
// go mod edit -replace voxwebsvr.com/client=./client
// go mod tidy

func SetRWriterStatus(w *http.ResponseWriter, code int) {
	(*w).WriteHeader(code)
}
func testCORS(w *http.ResponseWriter, r *http.Request) bool {
	header := (*w).Header()
	header.Set("Access-Control-Allow-Origin", "*")
	// header.Set("Access-Control-Allow-Credentials", "true")
	//header.Set("Access-Control-Allow-Headers", "Range, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
	if r.Method == "OPTIONS" {
		(*w).WriteHeader(http.StatusNoContent)
		return false
	}
	return true
}

var svrRootPath = "."

func handleRequest(w http.ResponseWriter, r *http.Request) {

	if testCORS(&w, r) {

		pathStr := svrRootPath + r.URL.Path
		client.ReceiveRequest(&w, r, &pathStr)
	}
}
func main() {

	// for i ,v := range os.Args {
	// 	fmt.Println(i, v)
	// }
	rootPath, err := os.Getwd()
	if err == nil {
		//svrRootPath = rootPath
		fmt.Println("rootPath: ", rootPath)
	}
	var portStr string = "9090"
	argsLen := len(os.Args)
	// fmt.Println("argsLen: ", argsLen)
	if argsLen > 1 {
		portStr = "" + os.Args[1]
		fmt.Println("init current port: ", portStr)
	}
	handler := http.HandlerFunc(handleRequest)

	http.Handle("/static/", handler)

	fmt.Println("cpus number: ", runtime.NumCPU())
	fmt.Println("Web Server version 1.0.12")
	fmt.Println("Web Server started at port: ", portStr)
	http.ListenAndServe(":"+portStr, nil)
}
