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
var ScheduleURL = "http://nagayo.sakura.ne.jp/cgi/schedule/schedule.cgi"

// LocationMap is a mapping between shortname and exact location
var LocationMap = map[string]string{
	"有楽町山野":    "ヤマノミュージックサロン有楽町 〒100-0006\\, 東京都千代田区\\, 有楽町2丁目10番1号",
	"スガナミ多摩":   "東京都多摩市落合1-46-1 ココリア多摩センター4F",
	"渋谷":       "東京都渋谷区神南1‐19‐4　日本生命アネックスビル５F",
	"リフラ":      "東京都新宿区新宿4-3-17 ダビンチ&",
	"★伊藤":      "〒272-0021 千葉県市川市八幡２丁目１５−10 パティオ 3階",
	"★フォーク有楽町": "ヤマノミュージックサロン有楽町 〒100-0006\\, 東京都千代田区\\, 有楽町2丁目10番1号",
}

// Nagayo responses iCal format data
func Nagayo(ctx *gin.Context) {
	day, timeStart, timeEnd, err := validateQuery(
		ctx.Query("day"),
		ctx.Query("start"),
		ctx.Query("end"),
	)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	vCalendar := VCalendar{
		Calname:  "石川永世 レッスンスケジュール",
		Caldesc:  ScheduleURL,
		Timezone: "Asia/Tokyo",
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

		events, err := Parse(doc, day)
		if err != nil {
			log.Println(err)
			continue
		}

		for i := range events {
			e := &events[i]

			e.Location = LocationMap[e.Summary]
			e.TimeStart = timeStart
			e.TimeEnd = timeEnd
			e.Tzid = vCalendar.Timezone
		}

		vCalendar.Events = append(vCalendar.Events, events...)
	}

	ctx.Header("Content-Type", "text/calendar")
	ctx.String(http.StatusOK, vCalendar.String())
}

func validateQuery(qDay string, qStart string, qEnd string) (day int, start string, end string, err error) {
	day, err = strconv.Atoi(qDay)
	if err != nil {
		err = fmt.Errorf(`"day" is missing or not an integer: "%s"`, qDay)
		return
	}
	if day < 0 || 7 < day {
		err = fmt.Errorf(`"day" is out of range (0-7): "%d"`, day)
		return
	}

	r := regexp.MustCompile(`^(?:[01][0-9]|2[0-3])[0-5][0-9]$`)
	if !r.MatchString(qStart) {
		err = fmt.Errorf(`"start" is missing or not valid pattern (HHMM, 24 hours): "%s"`, qStart)
		return
	}
	if !r.MatchString(qEnd) {
		err = fmt.Errorf(`"end" is missing or not valid pattern (HHMM, 24 hours): "%s"`, qEnd)
		return
	}

	return day, qStart, qEnd, nil
}

// VCalendar represents VCALENDAR in ics format
type VCalendar struct {
	Calname  string
	Caldesc  string
	Timezone string // example: Asia/Tokyo
	Events   []VEvent
}

// String returns a string of BEGIN:VCALENDAR...END:VCALENDAR format
func (c *VCalendar) String() string {
	var calendar string
	calendar += "BEGIN:VCALENDAR\n"
	calendar += "X-WR-CALNAME:" + c.Calname + "\n"
	calendar += "X-WR-CALDESC:" + c.Caldesc + "\n"
	calendar += "X-WR-TIMEZONE:" + c.Timezone + "\n"
	for _, e := range c.Events {
		calendar += e.String()
	}
	calendar += "END:VCALENDAR\n"

	return calendar
}

// VEvent represents VEVENT in ics format
type VEvent struct {
	Summary   string
	Location  string
	Date      string // example: 201909
	TimeStart string // example: 0930
	TimeEnd   string // example: 2305
	Tzid      string // example: Asia/Tokyo
}

// String returns a string of BEGIN:VEVENT...END:VEVENT format
func (e *VEvent) String() string {
	var event string
	event += "BEGIN:VEVENT\n"
	event += "SUMMARY:" + e.Summary + "\n"
	event += "LOCATION:" + e.Location + "\n"
	event += fmt.Sprintf("DTSTART;TZID=%s:%sT%s00", e.Tzid, e.Date, e.TimeStart) + "\n"
	event += fmt.Sprintf("DTEND;TZID=%s:%sT%s00", e.Tzid, e.Date, e.TimeEnd) + "\n"
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
			Summary: d.summary,
			Date:    fmt.Sprintf("%04s%02s%02s", ym.year, ym.month, d.date),
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
	var elems *goquery.Selection
	if day == 0 {
		allCol := "tr:not(:first-of-type) >td"
		elems = doc.Find(allCol)
	} else {
		specificCol := fmt.Sprintf("tr:not(:first-of-type) >td:nth-of-type(%d)", day)
		elems = doc.Find(specificCol)
	}

	elems.Each(func(i int, elem *goquery.Selection) {
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
