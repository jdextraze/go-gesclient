package main

import "fmt"

func main() {
	ch := make(chan bool, 1)
	ch <- true
	fmt.Println(<-ch)
}
