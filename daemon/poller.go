// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/melvinmt/firebase"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	numPollers        = 2                // number of Poller goroutines to launch
	pollInterval      = 10 * time.Second // how often to poll each URL
	statusInterval    = 10 * time.Second // how often to log status to stdout
	errTimeout        = 10 * time.Second // back-off timeout on error
	baseUrl           = "https://www.soccerfours.com"
	registerUrl       = baseUrl + "/index.php/component/dtregister/"
	selector          = ".event_message tr[class^=\"eventList\"][valign=\"top\"]"
	dateFmt           = "Jan _2, 2006 15:00 (MST)"
	firebaseUrl       = "https://fiery-heat-4051.firebaseio.com"
	firebaseAuthToken = "q8DjKMz5qM8QRARxynppt6XbvcbdhdetS0FgarmZ"
)

var (
	categories = []int{1}
	m          = make(map[int]Event)
	removed    = make(map[int]bool)
)

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

type ByDate []Event

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

// State represents the last-known state of a URL.
type State struct {
	category int
	status   string
}

// StateMonitor maintains a map that stores the state of the URLs being
// polled, and prints the current state every updateInterval nanoseconds.
// It returns a chan State to which resource state should be sent.
func StateMonitor(updateInterval time.Duration) chan<- State {
	updates := make(chan State)
	urlStatus := make(map[int]string)
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				logState(urlStatus)
			case s := <-updates:
				urlStatus[s.category] = s.status
			}
		}
	}()
	return updates
}

// logState prints a state map.
func logState(s map[int]string) {
	log.Println("Current state:")
	//for k, v := range s {
	//	log.Printf(" %d %s", k, v)
	//}
	log.Println("Number of events available:", len(m))
	for k := range m {
		removed[k] = true
	}

	for k, v := range m {
		refUrl := fmt.Sprintf("%s/events/%d", firebaseUrl, k)
		ref := firebase.NewReference(refUrl).Auth(firebaseAuthToken)
		if err := ref.Write(v); err != nil {
			log.Println(err, k)
		}
		delete(removed, k)
	}
	log.Println("Number of events removed  :", len(removed))
	for k, _ := range removed {
		refUrl := fmt.Sprintf("%s/events/%d", firebaseUrl, k)
		ref := firebase.NewReference(refUrl).Auth(firebaseAuthToken)
		if err := ref.Delete(); err != nil {
			log.Println(err, k)
		}
	}
}

// Resource represents an HTTP URL to be polled by this program.
type Resource struct {
	category int
	errCount int
}

// Poll executes an HTTP HEAD request for url
// and returns the HTTP status string or an error string.
func (r *Resource) Poll() string {
	data := url.Values{
		"category": {fmt.Sprintf("%d", r.category)},
		"task":     {"category"},
	}
	resp, err := http.PostForm(registerUrl, data)
	if err != nil {
		log.Println("Error", r.category, err)
		r.errCount++
		return err.Error()
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(io.Reader(resp.Body))
	if err != nil {
		log.Println("Error", r.category, err)
		r.errCount++
	}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		td := s.Find("td")

		aTags := td.First().Find("a")
		if aTags.Length() < 3 {
			return
		}
		n := strings.TrimSpace(aTags.Eq(0).Text())

		href := ""
		aTags.Each(func(i int, s *goquery.Selection) {
			if s.Text() == "Full" {
				href, _ = s.Attr("href")
			}
		})
		if href == "" {
			return
		}
		href = strings.Replace(href, "=register", "=individualRegister", 1)
		re := regexp.MustCompile("eventId=([0-9]+)")
		matched := re.FindStringSubmatch(href)
		evid, _ := strconv.ParseInt(matched[1], 10, 16)

		dStr := td.Eq(1).Find("span").Text()
		if strings.Contains(n, "7") {
			s := []string{dStr, "19:00", "(PST)"}
			dStr = strings.Join(s, " ")
		} else if strings.Contains(n, "8") {
			s := []string{dStr, "20:00", "(PST)"}
			dStr = strings.Join(s, " ")
		}

		loc, _ := time.LoadLocation("America/Los_Angeles")
		d, err := time.ParseInLocation(dateFmt, dStr, loc)
		if err != nil {
			r.errCount++
		}

		pStr := td.Eq(2).Text()
		cStr := td.Eq(3).Text()
		rStr := td.Eq(4).Text()
		aStr := td.Eq(5).Text()

		pStr = strings.Replace(pStr, ".", "", 1)
		pStr = strings.Replace(pStr, "$", "", 1)
		pStr = strings.TrimSpace(pStr)
		cStr = strings.TrimSpace(cStr)
		rStr = strings.TrimSpace(rStr)
		aStr = strings.TrimSpace(aStr)

		p, err := strconv.ParseInt(pStr, 10, 16)
		if err != nil {
		}
		c, err := strconv.ParseInt(cStr, 10, 8)
		if err != nil {
		}
		r, err := strconv.ParseInt(rStr, 10, 8)
		if err != nil {
		}
		a, err := strconv.ParseInt(aStr, 10, 8)
		if err != nil {
		}

		e := Event{
			Id:         int(evid),
			Name:       n,
			Date:       d.UTC(),
			Price:      int(p),
			Capacity:   int(c),
			Registered: int(r),
			Available:  int(a),
			Url:        baseUrl + href,
		}
		m[int(d.Unix())] = e
	})

	r.errCount = 0
	return "NAH"
}

// Sleep sleeps for an appropriate interval (dependent on error state)
// before sending the Resource to done.
func (r *Resource) Sleep(done chan<- *Resource) {
	time.Sleep(pollInterval + errTimeout*time.Duration(r.errCount))
	done <- r
}

func Poller(in <-chan *Resource, out chan<- *Resource, status chan<- State) {
	for r := range in {
		s := r.Poll()
		status <- State{r.category, s}
		out <- r
	}
}

func startPolling() {
	// Create our input and output channels.
	pending, complete := make(chan *Resource), make(chan *Resource)

	// Launch the StateMonitor.
	status := StateMonitor(statusInterval)

	// Launch some Poller goroutines.
	for i := 0; i < numPollers; i++ {
		go Poller(pending, complete, status)
	}

	// Send some Resources to the pending queue.
	go func() {
		for _, category := range categories {
			pending <- &Resource{category: category}
		}
	}()

	for r := range complete {
		go r.Sleep(pending)
	}
}

func main() {
	startPolling()
}
