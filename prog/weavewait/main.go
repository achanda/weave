package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	ErrNoCommandSpecified = errors.New("No command specified")
)

func main() {
	usr2 := make(chan os.Signal)
	signal.Notify(usr2, syscall.SIGUSR2)
	<-usr2

	args := os.Args[1:]

	if len(args) == 0 {
		checkErr(ErrNoCommandSpecified)
	}

	binary, err := exec.LookPath(args[0])
	checkErr(err)

	checkErr(syscall.Exec(binary, args, os.Environ()))
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
