package main

import "os"

func callExit() {
	os.Exit(1)
}

func main() {
	println("here is it")
	callExit()

	os.Exit(1) // want `os.Exit called in main/main`
}
