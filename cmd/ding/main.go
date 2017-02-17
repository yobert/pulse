package main

import (
	"fmt"
	"github.com/yobert/pulse"
)

func main() {
	if err := pulse.Ding(); err != nil {
		fmt.Println(err)
	}
}
