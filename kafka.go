package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

func InitKafkaProducer() *kafka.Producer {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create kafka producer")
	}
	return producer
}

func ProduceEvent(producer *kafka.Producer, topic string, msg []byte) {
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: msg,
	}, nil)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"topic":   topic,
			"message": string(msg),
		}).Error("Failed to produce event")
	} else {
		logrus.WithFields(logrus.Fields{
			"topic":   topic,
			"message": string(msg),
		}).Info("Successfully produced event")
	}
}
