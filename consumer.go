package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var currentTicketCount int

func InitKafkaConsumer(topics []string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"group.id":          "mygroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create kafka consumer")
	}

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to subscribe to topics")
	}

	return consumer
}

func ConsumeEvents(consumer *kafka.Consumer, topics []string) {
	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to read message")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"topic":   msg.TopicPartition.Topic,
			"message": string(msg.Value),
		}).Info("Received message")

		switch *msg.TopicPartition.Topic {
		case bookTicketEvent:
			// do something
			if currentTicketCount >= maxTicketCount {
				logrus.Info("Max ticket count reached")
				continue
			}
			currentTicketCount++
		case cancelTicketEvent:
			// do something
			if currentTicketCount <= 0 {
				logrus.Info("No ticket available")
				continue
			}
			currentTicketCount--
		}
		logrus.Info("Current ticket count: ", currentTicketCount)
	}
}
