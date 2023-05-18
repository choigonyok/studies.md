package main

import "fmt"

func before() {
	fmt.Println("I am before")
}

func after() {
	fmt.Println("I am after")
}

func main() {
	defer before()
	defer after()
}
