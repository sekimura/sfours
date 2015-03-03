package main

import "time"

type Event struct {
	Id         int       `json:"id"`
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	Price      int       `json:"price"`
	Capacity   int       `json:"capacity"`
	Registered int       `json:"registered"`
	Available  int       `json:"available"`
	Url        string    `json:"url"`
}
