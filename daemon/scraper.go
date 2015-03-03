package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	selector = ".event_message tr[class^=\"eventList\"][valign=\"top\"]"
	dateFmt  = "Jan _2, 2006 15:00 (MST)"
)

func ScrapeFromResponse(res *http.Response) (events []Event, err error) {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Println("Error", err)
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
			} else if s.Text() == "Register" {
				href, _ = s.Attr("href")
			}
		})
		if href == "" {
			return
		}
		href = strings.Replace(href, "=register", "=individualRegister", 1)
		log.Println(href, n)
		re := regexp.MustCompile("eventId=([0-9]+)")
		matched := re.FindStringSubmatch(href)
		evid, _ := strconv.ParseInt(matched[1], 10, 16)

		dStr := td.Eq(1).Find("span").Text()
		if strings.Contains(n, "7pm") {
			s := []string{dStr, "19:00", "(PST)"}
			dStr = strings.Join(s, " ")
		} else if strings.Contains(n, "8pm") {
			s := []string{dStr, "20:00", "(PST)"}
			dStr = strings.Join(s, " ")
		}
		d, err := time.Parse(dateFmt, dStr)
		if err != nil {
			log.Println("Error", err)
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

		events = append(events, Event{
			Id:         int(evid),
			Name:       n,
			Date:       d,
			Price:      int(p),
			Capacity:   int(c),
			Registered: int(r),
			Available:  int(a),
			Url:        baseUrl + href,
		})
	})

	return
}
