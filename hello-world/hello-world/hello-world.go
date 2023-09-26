// Copyright 2022 Hewlett Packard Enterprise Development LP

package main

import "fmt"

func Hello() string {
	var s string
    s = "hello world"
	return s
}

func main() {
	fmt.Println(Hello())
}
