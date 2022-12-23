package main

import "fmt"

func main() {
}

//export handle
func handle() {
	fmt.Println("Content-Type: text/plain")
	fmt.Println("Status: 200")
	fmt.Println("")
	fmt.Println("hello world from tinygo handler")
}
