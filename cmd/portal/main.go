package main

import (
	"breaker/cmd/portal/command"
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
