package main

import (
	"fmt"
	"os"

	"github.com/impadalko/CES27Projeto/sign"
)

func main() {
	err := sign.TestWriteAndReadPemFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sign.TestSignAndVerify()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ALL TESTS PASSED")
}