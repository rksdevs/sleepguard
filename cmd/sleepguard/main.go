package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	// add command-line flags for device name, alert cooldown, and debug mode
	device := flag.String("device", "laptop", "name of the device")
	alertInterval := flag.Duration("alertInterval",30*time.Second, "default alter interval")
	debug := flag.Bool("debug", true, "are we debugging")
	
	flag.Parse()

	fmt.Println("device: ", *device)
	fmt.Println("alertInterval: ", *alertInterval)
	fmt.Println("debug: ", *debug)

	fmt.Println("here we go")
}