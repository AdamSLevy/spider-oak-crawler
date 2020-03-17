package main

import "fmt"

func list() error {
	fmt.Println("list")
	return nil
}
func start(url string) error {
	fmt.Println("start", url)
	return nil
}
func stop(url string) error {
	fmt.Println("stop", url)
	return nil
}
