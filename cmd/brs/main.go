package main

import (
	"fmt"
	"os"
)

func main() {
	err := cmdRoot.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
