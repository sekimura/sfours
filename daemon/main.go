package main

import "log"

func main() {
	webFetcherChan := FetcherGen()
	firebasePublisherChan := PublisherGen(webFetcherChan)
	for event := range firebasePublisherChan {
		if event.Id > 0 {
			log.Println("YO", event)
		}
	}
}
