package main

import (
	"fmt"
	"github.com/domgolonka/proxy-scraper"
	"time"
)

func main() {
	gen := proxy.New(1, 5*time.Second, 5, nil)
	time.Sleep(10 * time.Second)
	gen.Run(1, 1)
	fmt.Println(gen.Count())
	fmt.Println(gen.GetProxies())
}
