package main

import (
	"fmt"
	"template-go/pkg/kyden"
	"template-go/pkg/tmutil"
)

const s = "Kyden"

func main() {
	fmt.Println("This is a Golang Template Project")

	k := kyden.NewKyden(s)
	k.Say(s)
	println(k.Hello())

	fmt.Println(tmutil.CurrDate())
	fmt.Println(tmutil.CurrMonth())
	fmt.Println(tmutil.CurrYear())
	fmt.Println(tmutil.CurrHour())
	fmt.Println(tmutil.CurrWeek())
	fmt.Println(tmutil.CurrWeek2())

	fmt.Printf("k.Name(): %v\n", k.Name())
}
