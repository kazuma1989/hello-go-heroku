package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// ScheduleURL is URL to parse
const ScheduleURL = "http://nagayo.sakura.ne.jp/cgi/schedule/schedule.cgi"

// Nagayo responses iCal format data
func Nagayo(ctx *gin.Context) {
	vCalendar := VCalendar{
		calname:  "石川永世 レッスンスケジュール",
		caldesc:  ScheduleURL,
		timezone: "Asia/Tokyo",
	}

	now := time.Now()
	for _, t := range []time.Time{
		now.AddDate(0, -1, 0), // prev month
		now,                   // this month
		now.AddDate(0, 1, 0),  // next month
	} {
		query := url.Values{}
		query.Set("year", strconv.Itoa(t.Year()))
		query.Set("month", strconv.Itoa(int(t.Month())))

		doc, err := goquery.NewDocument(ScheduleURL + "?" + query.Encode())
		if err != nil {
			log.Println(err)
			continue
		}

		day, err := strconv.Atoi(ctx.Query("day"))
		if err != nil {
			log.Println(err)
			continue
		}
		if day <= 0 || 8 <= day {
			log.Printf(`"day" is out of range (1-7): %d`, day)
			continue
		}

		vCalendar.events, err = Parse(doc, day)
		if err != nil {
			log.Println(err)
			continue
		}

		for i := range vCalendar.events {
			vCalendar.events[i].location = "ヤマノミュージックサロン有楽町 〒100-0006\\, 東京都千代田区\\, 有楽町2丁目10番1号"
			vCalendar.events[i].timeStart = "1930"
			vCalendar.events[i].timeEnd = "2030"
			vCalendar.events[i].tzid = vCalendar.timezone
		}
	}

	ctx.Header("Content-Type", "text/calendar")
	ctx.String(http.StatusOK, vCalendar.String())
}

// VCalendar represents VCALENDAR in ics format
type VCalendar struct {
	calname  string
	caldesc  string
	timezone string // example: Asia/Tokyo
	events   []VEvent
}

// String returns a string of BEGIN:VCALENDAR...END:VCALENDAR format
func (c *VCalendar) String() string {
	var calendar string
	calendar += "BEGIN:VCALENDAR\n"
	calendar += "X-WR-CALNAME:" + c.calname + "\n"
	calendar += "X-WR-CALDESC:" + c.caldesc + "\n"
	calendar += "X-WR-TIMEZONE:" + c.timezone + "\n"
	for _, e := range c.events {
		calendar += e.String()
	}
	calendar += "END:VCALENDAR\n"

	return calendar
}

// VEvent represents VEVENT in ics format
type VEvent struct {
	summary   string
	location  string
	date      string // example: 201909
	timeStart string // example: 0930
	timeEnd   string // example: 2305
	tzid      string // example: Asia/Tokyo
}

// String returns a string of BEGIN:VEVENT...END:VEVENT format
func (e *VEvent) String() string {
	var event string
	event += "BEGIN:VEVENT\n"
	event += "SUMMARY:" + e.summary + "\n"
	event += "LOCATION:" + e.location + "\n"
	event += fmt.Sprintf("DTSTART;TZID=%s:%sT%s00", e.tzid, e.date, e.timeStart) + "\n"
	event += fmt.Sprintf("DTEND;TZID=%s:%sT%s00", e.tzid, e.date, e.timeEnd) + "\n"
	event += "END:VEVENT\n"

	return event
}

// Parse parses the schedule page
func Parse(doc *goquery.Document, day int) (events []VEvent, err error) {
	ym, err := parseYearMonth(doc)
	if err != nil {
		log.Println(err)
		return
	}

	dateCells, err := parseDate(doc, day)
	if err != nil {
		log.Println(err)
		return
	}

	for _, d := range dateCells {
		events = append(events, VEvent{
			summary: d.summary,
			date:    fmt.Sprintf("%04s%02s%02s", ym.year, ym.month, d.date),
		})
	}

	return events, nil
}

type yearMonthCell struct {
	year  string
	month string
}

func parseYearMonth(doc *goquery.Document) (cell yearMonthCell, err error) {
	innerText := ConvertEUCJP(doc.Find("td[colspan='7']").Text())

	// 2019年2月 -> [2019年2月, 2019, 2]
	r := regexp.MustCompile(`^([0-9]{4}).*?([0-9]{1,2}).*$`)
	match := r.FindStringSubmatch(innerText)
	parseLog.Println(match)

	if len(match) != 3 {
		err = fmt.Errorf("No match found in: %s", innerText)
		return
	}

	return yearMonthCell{
		year:  match[1],
		month: match[2],
	}, nil
}

type dateCell struct {
	date    string
	summary string
}

func parseDate(doc *goquery.Document, day int) (cells []dateCell, err error) {
	doc.Find(fmt.Sprintf("tr:not(:first-of-type) >td:nth-of-type(%d)", day)).Each(func(i int, elem *goquery.Selection) {
		innerText := ConvertEUCJP(elem.Text())

		// 12有楽町山野 -> [12有楽町山野, 12, 有楽町山野]
		r := regexp.MustCompile(`^([0-9]{1,2})(.*$)`)
		match := r.FindStringSubmatch(innerText)
		parseLog.Println(match)

		if len(match) != 3 {
			return
		}

		date := match[1]
		summary := match[2]
		if summary == "" {
			return
		}

		cells = append(cells, dateCell{date, summary})
	})

	return cells, nil
}
