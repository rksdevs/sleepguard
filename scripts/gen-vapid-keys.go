// One-off helper to print a VAPID key pair for deploy/.env
package main

import (
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	private, public, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		panic(err)
	}
	fmt.Println("SLEEPGUARD_VAPID_PUBLIC_KEY=" + public)
	fmt.Println("SLEEPGUARD_VAPID_PRIVATE_KEY=" + private)
}
