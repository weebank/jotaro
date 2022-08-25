package main

import (
	"flag"

	"github.com/weebank/jotaro/test/consumer"
	"github.com/weebank/jotaro/test/producer"
)

func main() {
	produce := flag.Bool("p", false, "send message instead of receive")
	flag.Parse()

	// Run producer or consumer main function
	if *produce {
		producer.Main()
	} else {
		consumer.Main()
	}
}
