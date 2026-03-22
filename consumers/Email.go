package consumers

import (
	"fmt"
	"time"
)

func MailConsumer() {
	for {
		time.Sleep(3 * time.Second)

		fmt.Println("Consumer: menerima message dari queue")
	}
}
