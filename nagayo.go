package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// ScheduleURL is URL to parse
const ScheduleURL = "http://nagayo.sakura.ne.jp/cgi/schedule/schedule.cgi"

// Nagayo responses iCal format data
func Nagayo(ctx *gin.Context) {
	now := time.Now()

	var events []string
	for _, ym := range []time.Time{
		now.AddDate(0, -1, 0), // prev month
		now,                   // this month
		now.AddDate(0, 1, 0),  // next month
	} {
		query := url.Values{}
		query.Set("year", strconv.Itoa(ym.Year()))
		query.Set("month", strconv.Itoa(int(ym.Month())))

		doc, err := goquery.NewDocument(ScheduleURL + "?" + query.Encode())
		if err != nil {
			log.Println(err)
			continue
		}

		parsed, err := Parse(doc)
		if err != nil {
			log.Println(err)
			continue
		}

		events = append(events, parsed...)
	}

	var calendar string
	calendar += "BEGIN:VCALENDAR\n"
	calendar += "X-WR-CALNAME:石川永世 レッスンスケジュール\n"
	calendar += "X-WR-CALDESC:" + ScheduleURL + "\n"
	calendar += "X-WR-TIMEZONE:Asia/Tokyo\n"
	calendar += strings.Join(events, "")
	calendar += "END:VCALENDAR\n"

	ctx.Header("Content-Type", "text/calendar")
	ctx.String(http.StatusOK, calendar)
}

// Parse parses the schedule page
func Parse(doc *goquery.Document) (events []string, err error) {
	ym, err := parseYearMonth(doc)
	if err != nil {
		log.Println(err)
		return
	}

	dateCells, err := parseDate(doc, 3)
	if err != nil {
		log.Println(err)
		return
	}

	for _, d := range dateCells {
		event := "BEGIN:VEVENT\n"
		event += "SUMMARY:" + d.summary + "\n"
		event += "LOCATION:ヤマノミュージックサロン有楽町 〒100-0006\\, 東京都千代田区\\, 有楽町2丁目10番1号\n"
		event += "DTSTART;TZID=Asia/Tokyo:" + fmt.Sprintf("%04s%02s%02sT193000", ym.year, ym.month, d.date) + "\n"
		event += "DTEND;TZID=Asia/Tokyo:" + fmt.Sprintf("%04s%02s%02sT203000", ym.year, ym.month, d.date) + "\n"
		event += "END:VEVENT\n"

		events = append(events, event)
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
