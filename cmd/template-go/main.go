package main

import (
	"fmt"
	"template-go/pkg/kyden"
	"time"
)

const (
	errno = -6398
)

func main() {
	va := make([]int, 0)
	fmt.Println(len(va))

	fmt.Println(errno)

	println(time.Now().Format(time.DateTime))

	pk := kyden.NewKyden("k")
	fmt.Println(pk.Hello())
}
