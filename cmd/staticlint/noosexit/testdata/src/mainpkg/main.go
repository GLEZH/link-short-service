package main

import (
	. "os"
	operatingsystem "os"
)

type terminator struct{}

func (terminator) Exit(int) {}

func main() {
	operatingsystem.Exit(1) // want "direct call to os.Exit in main is prohibited"
	Exit(1)                 // want "direct call to os.Exit in main is prohibited"

	var os terminator
	os.Exit(1)
}

func helper() {
	operatingsystem.Exit(1)
}
