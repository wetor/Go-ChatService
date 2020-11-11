package rabbitmq

import (
	"fmt"
	"strconv"

	"github.com/streadway/amqp"
)

// Test 测试函数，接收函数。
func Test() {
	rabbitmq := NewRabbitMQ("testProduct")

	rabbitmq.PublishPub("{这是测试用的}")
	forever := make(chan bool)
	msgID := 0
	rabbitmq.ReceiveSub(func(message <-chan amqp.Delivery) {
		for d := range message {
			msg := string(d.Body)
			fmt.Printf("ConsumerC Receive: %s\n", msg)
			if msg == "End" {
				forever <- true
				return
				//os.Exit(1)
			}
			fmt.Println("test" + strconv.Itoa(msgID))
			msgID++

		}
	})

	<-forever
	fmt.Println("end")
}
