package main

import (
	"fmt"
	"log"

	"github.com/melvinmt/firebase"
)

const (
	firebaseUrl       = "https://fiery-heat-4051.firebaseio.com"
	firebaseAuthToken = "q8DjKMz5qM8QRARxynppt6XbvcbdhdetS0FgarmZ"
)

func PublisherGen(inChan chan Event) chan Event {
	out := make(chan Event)
	go func() {
		for event := range inChan {
			k := event.Date.Unix()
			refUrl := fmt.Sprintf("%s/events/%d", firebaseUrl, k)
			ref := firebase.NewReference(refUrl).Auth(firebaseAuthToken)
			if err := ref.Write(event); err != nil {
				log.Println(err, k)
			} else {
				out <- event
			}
		}
	}()
	return out
}
