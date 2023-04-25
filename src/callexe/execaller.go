package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("go exe caller Hello World!")
	datapath := "./exeForGo.exe ver=1.0 --help"
	cmd := exec.Command("cmd.exe", "/c", "start "+datapath)
	cmd.Run()
}
