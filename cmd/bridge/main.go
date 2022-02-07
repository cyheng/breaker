package main

import (
	"breaker/cmd/bridge/command"
	"fmt"
	"os"
)

func main() {
	err := command.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
