1. 安装 golang 语言环境
2. 安装 gin框架, 细节请见 https://blog.csdn.net/vily_lei/article/details/125695689
3. 在httpserver/websvr/client目录里面建立mod文件：在此目录下执行 go mod init voxwebsvr.com/client 命令即可
4. 在httpserver/websvr/webfs建立mod文件：在此目录下执行 go mod init voxwebsvr.com/webfs 命令即可
5. 在httpserver/websvr/目录里面建立mod文件：
    step1. 在websvr/目录里执行 go mod init voxwebsvr.com/main 命令
    step2. 接着执行 go mod edit -replace voxwebsvr.com/webfs=./webfs 命令
    step3. 接着执行 go mod edit -replace voxwebsvr.com/client=./client 命令
    step2. 最后执行 go mod tidy 命令
6. 在httpserver/websvr/目录里编译: 在此目录里执行 go build -o ../../bin/ httpserver.go 命令
   就会在 httpserver/bin/ 目录里生成对应平台的可执行程序