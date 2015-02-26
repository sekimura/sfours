package main

import "time"

type Event struct {
	Id         int
	Name       string
	Date       time.Time
	Price      int
	Capacity   int
	Registered int
	Available  int
	Url        string
}
