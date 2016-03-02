package main

import (
	"fmt"
//	"runtime"
	"time"
)

var i int = 0

func func1(){
	for j := 0; j < 1e6; j++{
		i += 1
	}
}

func func2(){
	for j := 0; j < 1e6; j++{
		i -= 1
	}
}

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	go func1()
	go func2()
	time.Sleep(100*time.Millisecond)
	fmt.Println(i)
}
