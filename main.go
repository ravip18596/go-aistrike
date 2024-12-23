package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var baseEvent Event
var kafkaProducer *kafka.Producer

func heartbeatHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(heartbeatResponse{
		Status: "OK",
		Code:   http.StatusOK,
	})
}

// Middleware should be able to handle panic
func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//Middleware should be able to handle panic
				logrus.WithFields(logrus.Fields{
					"URL": r.RequestURI,
				}).Errorf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func handler() http.Handler {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.Use(recoverHandler)
	router.HandleFunc("/", heartbeatHandlerFunc)
	router.HandleFunc("/tickets", bookTicket).Methods("POST")
	router.HandleFunc("/tickets/{id}", cancelTicket).Methods("DELETE")
	router.HandleFunc("/users/{user_id}/tickets", getUserBookings).Methods("GET")

	return router
}

func CreateTables() {
	query := `CREATE TABLE IF NOT EXISTS events(
    					id INT AUTO_INCREMENT PRIMARY KEY,
    					name VARCHAR(255) NOT NULL
              );`
	_, err := mysql.Exec(query)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"query": query,
		}).Error("Unable to create table events")
	} else {
		logrus.WithFields(logrus.Fields{
			"query": query,
		}).Info("Created table events")
	}

	query = `CREATE TABLE IF NOT EXISTS users(
    					id INT AUTO_INCREMENT PRIMARY KEY,
    					name VARCHAR(255) NOT NULL
              );`
	_, err = mysql.Exec(query)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"query": query,
		}).Error("Unable to create table users")
	} else {
		logrus.WithFields(logrus.Fields{
			"query": query,
		}).Info("Created table users")
	}

	query = `CREATE TABLE IF NOT EXISTS bookings(
    					id INT AUTO_INCREMENT PRIMARY KEY,
						user_id INT NOT NULL,
						event_id INT NOT NULL,
						booking_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_id) REFERENCES users(id),
						FOREIGN KEY (event_id) REFERENCES events(id),
						UNIQUE KEY unique_booking (user_id, event_id)
              );`
	_, err = mysql.Exec(query)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"query": query,
		}).Error("Unable to create table bookings")
	} else {
		logrus.WithFields(logrus.Fields{
			"query": query,
		}).Info("Created table bookings")
	}
}

func CreateBaseEvent() {
	eventName := "Diljit Concert"
	eventId := 1
	if mysql == nil {
		mysql = InitMysql()
	}
	insertStm, err := mysql.Prepare("insert into events(id, name) values (?, ?);")
	if err != nil {
		logrus.Error("mysql prepare statement error", err)
	}
	defer insertStm.Close()
	_, err = insertStm.Exec(eventId, eventName)
	if err != nil {
		logrus.Error("mysql prepare statement execution error", err)
	}
	baseEvent.Id = eventId
	baseEvent.Name = eventName
	logrus.WithFields(logrus.Fields{
		"id":   baseEvent.Id,
		"name": baseEvent.Name,
	}).Info("Base Event Created")
}

func main() {
	logrus.Info("starting aistrike web server")
	time.Sleep(45 * time.Second)
	go func() {
		CreateTables()
		CreateBaseEvent()
		topics := []string{bookTicketEvent, cancelTicketEvent}
		kafkaConsumer := InitKafkaConsumer(topics)

		ConsumeEvents(kafkaConsumer, topics)
	}()
	kafkaProducer = InitKafkaProducer()
	logrus.Info("starting http web server started at port 8080")
	errHttp := http.ListenAndServe(":8080", handler())
	if errHttp != nil {
		logrus.WithFields(logrus.Fields{
			"addr": "3001",
			"err":  errHttp,
		}).Fatal("Unable to create HTTP Service ")
	}
}
