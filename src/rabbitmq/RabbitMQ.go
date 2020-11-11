// 函数说明：https://blog.csdn.net/vrg000/article/details/81165030
// 代码来源：http://www.topgoer.com/数据库操作/go操作RabbitMQ/Publish模式.html
// 多为固定写法

package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// MQURL 连接信息
const MQURL = "amqp://wetor:123456@192.168.10.104:5672"

//RabbitMQ rabbitMQ类
type RabbitMQ struct {
	//连接
	conn *amqp.Connection
	//通道
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机名称
	Exchange string
	//bind key
	Key string

	//public static RabbitMQ NewRabbitMQPubSub(exchangeName string)

	//public void Destroy()

	//private void failOnErr(err error, message string)

	//public void PublishPub(message string)

	//public void ReceiveSub()
}

//NewRabbitMQ 订阅模式-创建RabbitMQ实例
func NewRabbitMQ(exchangeName string) *RabbitMQ {
	//创建RabbitMQ实例
	rabbitmq := &RabbitMQ{QueueName: "", Exchange: exchangeName, Key: ""}
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(MQURL)
	rabbitmq.failOnErr(err, "连接RabbitMQ失败!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "打开Channel失败!")
	return rabbitmq
}

//Destroy 断开channel和connection
func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

//failOnErr 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

//PublishPub 订阅模式-生产者 *固定写法
func (r *RabbitMQ) PublishPub(message string) {
	//1.尝试创建交换机
	//一直保存在redis中不会自动删除
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout", //订阅模式
		true,     //持久性
		false,    //自动删除
		false,
		false,
		nil,
	)
	r.failOnErr(err, "声明(Declare)Exchange失败")

	//2.发送消息
	err = r.channel.Publish(
		r.Exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

/*
ReceiveSub 订阅模式-消费者 *固定写法

*/
func (r *RabbitMQ) ReceiveSub(callback func(message <-chan amqp.Delivery)) {
	//1.创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "声明(Declare)Exchange失败")
	//2.创建队列，这里注意队列名称不要写
	q, err := r.channel.QueueDeclare(
		"", //随机产生队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "声明(Declare)Queue失败")

	//绑定队列到exchange 中
	err = r.channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		"",
		r.Exchange,
		false,
		nil,
	)

	//消费消息
	message, err := r.channel.Consume(
		q.Name,
		"",
		true, //自动回复ACK
		false,
		false,
		false,
		nil,
	)
	// 构建一个通道
	//forever := make(chan bool)

	// 开启一个并发匿名函数
	go func() {
		go callback(message)
		//执行回调函数
	}()
	//<-forever // 从forever信道中取数据，必须要有数据流进来才可以，不然main在此阻塞
}
