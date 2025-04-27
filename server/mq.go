package main

import (
	"fmt"
	"log"
	"sync"

	"context"
	"encoding/json"
	"strconv"
	"time"

	"os"

	"aurora-im/common"

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

// 创建群exchange
func (mq *RabbitMQ) AddUser2Group(gid int64, uids ...int64) error {
	os.MkdirAll("data/im/server/group/", os.ModePerm)

	fout, err := os.OpenFile("data/im/server/group/"+strconv.FormatInt(gid, 10), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	for _, uid := range uids {
		fout.WriteString(strconv.FormatInt(uid, 10) + "\n")
	}
	fout.Close()

	group := "g" + strconv.FormatInt(gid, 10)
	exchange := group
	err = mq.mqChannel.ExchangeDeclare(
		exchange,
		"fanout", //type
		true,     //durable
		false,    //auto delete
		false,    //internal
		false,    //no-wait
		nil,      //arguments
	)
	if err != nil {
		return err
	}

	// 声明队列
	for _, uid := range uids {
		user := group + "_" + strconv.FormatInt(uid, 10)
		queues := []string{user + "_computer", user + "_mobile"}
		for _, QueueName := range queues {
			_, err = mq.mqChannel.QueueDeclare(
				QueueName,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				return err
			}

			// 绑定队列到exchange
			err = mq.mqChannel.QueueBind(
				QueueName, //Queue Name
				"",        //routing key。fout模式下会忽略routing key
				exchange,
				false, //noWait
				nil,   //arguments
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 向MQ里发送消息
func (mq *RabbitMQ) Send(message *common.Message, exchange string) error {
	// json序列化
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 	指定超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 向MQ发送消息
	err = mq.mqChannel.PublishWithContext(
		ctx,
		exchange,
		"",    //routing key
		false, //mandatory
		false, //immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json", //MIME content type
			Body:         msg,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// 释放MQ连接
func (mq *RabbitMQ) Release() {
	if mq.mqChannel != nil {
		mq.mqChannel.Close()
	}
	if mq.mqConn != nil {
		mq.mqConn.Close()
	}
}
