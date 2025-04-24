package main

import (
	"fmt"
	"log"
	"sync"

	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	mqConn    *amqp.Connection
	mqChannel *amqp.Channel //支持并发使用
}

var (
	mq     *RabbitMQ
	mqOnce sync.Once
)

func GetRabbitMQ() *RabbitMQ {
	mqOnce.Do(func() {
		conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", "guest", "guest"))
		if err != nil {
			log.Fatal(err)
		}

		ch, err := conn.Channel()
		if err != nil {
			log.Fatal(err)
		}

		mq = &RabbitMQ{
			mqConn:    conn,
			mqChannel: ch,
		}
	})

	return mq
}

// 创建user exchange 创建两个队列，然后和exchange绑定
func (mq *RabbitMQ) RegisterUser(uid int64, userType string) error {
	user := userType + strconv.FormatInt(uid, 10)
	exchange := user
	err := mq.mqChannel.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	queues := []string{
		user + "_computer",
		user + "_mobile",
	}

	for _, queue := range queues {
		_, err = mq.mqChannel.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}
		err = mq.mqChannel.QueueBind(
			queue,
			"", // fanout模式下，routingKey为空t模式下，routingKey为空
			exchange,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
