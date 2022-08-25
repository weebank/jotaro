package main

import (
	"flag"

	"github.com/weebank/jotaro/test/consumer"
	"github.com/weebank/jotaro/test/producer"
)

func main() {
	produce := flag.Bool("p", false, "send message instead of receive")
	count := flag.Uint("n", 10, "number of pokémons to send")
	flag.Parse()

	// Run producer or consumer main function
	if *produce {
		producer.Main(*count)
	} else {
		consumer.Main()
	}
}
