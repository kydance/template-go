package main

import (
	"fmt"
	"template-go/pkg/kyden"
)

const s = "Kyden"

func main() {
	fmt.Println("This is a Golang Template Project")

	kyden.Hello(s)
}
