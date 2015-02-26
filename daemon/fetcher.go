package main

import (
	"log"
	"net/http"
	"net/url"
)

type Request struct {
	Url    string
	Values map[string]string
}

const (
	baseUrl     = "https://www.soccerfours.com"
	registerUrl = baseUrl + "/index.php/component/dtregister/"
	category    = "1"
	task        = "category"
)

// Fetch web contents from soccerfours.com registration pages
// append "Event" to firebasePublisherChan
func FetcherGen() chan Event {
	out := make(chan Event)
	go func() {
		data := url.Values{
			"category": {category},
			"task":     {task},
		}
		res, err := http.PostForm(registerUrl, data)
		if err != nil {
			log.Fatal(err)
			return
		}
		events, err := ScrapeFromResponse(res)
		if err != nil {
			log.Fatal(err)
			return
		}
		if len(events) == 0 {
			var e Event
			out <- e
		} else {
			for _, v := range events {
				out <- Event(v)
			}
		}
	}()
	log.Println("out")
	return out
}
