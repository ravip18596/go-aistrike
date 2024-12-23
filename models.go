package main

import "time"

type heartbeatResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}

type Event struct {
	Id   int    `json:"event_id"`
	Name string `json:"event_name"`
}

type User struct {
	Id   int    `json:"user_id"`
	Name string `json:"user_name"`
}

type Bookings struct {
	Id          int       `json:"booking_id"`
	UserId      int       `json:"user_id"`
	EventId     int       `json:"event_id"`
	BookedTime  time.Time `json:"-"`
	BookingTime string    `json:"booking_time"`
}

type CancelEventBooking struct {
	Id      int `json:"booking_id"`
	UserId  int `json:"user_id"`
	EventId int `json:"event_id"`
}
