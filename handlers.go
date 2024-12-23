package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func bookTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if mysql == nil {
		mysql = InitMysql()
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		switch {
		case err.Error() == "http: request body too large":
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		default:
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		}
		return
	}
	//check if this user_id exists or not
	rows, err := mysql.Query("select id from users where id = ?;", u.Id)
	if err != nil {
		logrus.Error("mysql query statement error", err)
	}
	defer rows.Close()
	var cnt int
	for rows.Next() {
		cnt++
	}
	if cnt == 0 {
		// insert this user into mysql
		insertStm, err := mysql.Prepare("insert into users(id, name) values (?, ?);")
		if err != nil {
			logrus.Error("mysql prepare statement error", err)
		}
		defer insertStm.Close()
		_, err = insertStm.Exec(u.Id, u.Name)
		if err != nil {
			logrus.Error("mysql prepare statement execution error", err)
		}
		// insert it into redis
		key := fmt.Sprintf("user-%d", u.Id)
		valB, _ := json.Marshal(u)
		val := string(valB)
		err = redisClient.Set(ctx, key, val, 0).Err()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"key":   key,
			}).Error("redis set statement error")
		}
	}

	insertStm, err := mysql.Prepare("insert into bookings(user_id, event_id) values (?, ?);")
	if err != nil {
		logrus.Error("mysql prepare statement error", err)
	}
	defer insertStm.Close()
	_, err = insertStm.Exec(u.Id, baseEvent.Id)
	if err != nil {
		logrus.Error("mysql prepare statement execution error", err)
	}

	rows, err = mysql.Query("select id from bookings where user_id = ? and event_id = ?;", u.Id, baseEvent.Id)
	if err != nil {
		logrus.Error("mysql query statement error", err)
	}
	defer rows.Close()
	var bookingId int
	for rows.Next() {
		_ = rows.Scan(&bookingId)
	}
	booking := Bookings{
		Id:      bookingId,
		UserId:  u.Id,
		EventId: baseEvent.Id,
	}
	// inserting booking data into redis
	key := fmt.Sprintf("bookings-%d", bookingId)
	valB, _ := json.Marshal(booking)
	val := string(valB)
	err = redisClient.Set(ctx, key, val, 0).Err()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"key":   key,
		}).Error("redis set statement error")
	}
	// send event to kafka
	ProduceEvent(kafkaProducer, bookTicketEvent, valB)

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(booking)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error json marshal response")
	}
}

func cancelTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	bookingId, aErr := strconv.Atoi(vars["id"])
	if aErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := make(map[string]string)
		response["error"] = "Invalid Request"
		_ = json.NewEncoder(w).Encode(response)
	}
	if mysql == nil {
		mysql = InitMysql()
	}
	deleteStm, err := mysql.Prepare("delete from bookings where id = ?;")
	if err != nil {
		logrus.Error("mysql prepare statement error", err)
	}
	defer deleteStm.Close()
	_, err = deleteStm.Exec(bookingId)
	if err != nil {
		logrus.Error("mysql prepare statement execution error", err)
	} else {
		// delete booking from redis also
		key := fmt.Sprintf("bookings-%d", bookingId)
		var bookEvent Bookings
		var val string
		err = redisClient.Get(ctx, key).Scan(&val)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"key":   key,
			}).Error("redis get statement error")
		}
		err = json.Unmarshal([]byte(val), &bookEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"val":   val,
			}).Error("json unmarshal error")
		}

		err = redisClient.Del(ctx, key).Err()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"key":   key,
			}).Error("redis delete statement error")
		}
		// send event to kafka
		cancelTicket := CancelEventBooking{
			Id:      bookingId,
			UserId:  bookEvent.UserId,
			EventId: baseEvent.Id,
		}
		valB, _ := json.Marshal(cancelTicket)
		ProduceEvent(kafkaProducer, cancelTicketEvent, valB)
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]string)
	res["status"] = "success"
	res["message"] = fmt.Sprintf("Cancelled Ticket with id: %d", bookingId)
	_ = json.NewEncoder(w).Encode(res)
}

func getUserBookings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if mysql == nil {
		mysql = InitMysql()
	}
	vars := mux.Vars(r)
	userId, aErr := strconv.Atoi(vars["user_id"])
	if aErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := make(map[string]string)
		response["error"] = "Invalid Request"
		_ = json.NewEncoder(w).Encode(response)
	}
	rows, err := mysql.Query("select id, user_id, event_id, booking_time from bookings where user_id = ?;", userId)
	if err != nil {
		logrus.Error("mysql query statement error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		response := make(map[string]string)
		response["error"] = err.Error()
		_ = json.NewEncoder(w).Encode(response)
	}
	defer rows.Close()
	var results []Bookings
	for rows.Next() {
		var b Bookings
		err := rows.Scan(&b.Id, &b.UserId, &b.EventId, &b.BookingTime)
		if err != nil {
			logrus.Error("Error scanning row: ", err)
			continue
		}
		results = append(results, b)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error json marshal response")
	}
}
