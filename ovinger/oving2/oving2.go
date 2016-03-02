package main

import (
	"fmt"
    "runtime"
)

func func1(c chan int, d chan int){
	for j := 0; j < 1e6; j++{
		i := <-c + 1
		c <- i
	}
	d <- 1
}

func func2(c chan int, d chan int){
	for j := 0; j < 1e6; j++{
		i := <- c-1
		c <- i
	}
	d <-2
}

func main() {

	value := make(chan int,1)
	value <- 0
	sync := make(chan int, 2)

	go func1(value,sync)
	go func2(value,sync)
	
	<-sync
	<-sync
	fmt.Println(<-value)
}
