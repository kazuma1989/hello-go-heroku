package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var parseLog *log.Logger

func init() {
	parseLog = log.New(os.Stdout, "[Parse]", 0)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()
	router.LoadHTMLGlob("*.html")

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	})

	router.GET("/nagayo.ics", func(ctx *gin.Context) {
		doc, err := goquery.NewDocument("http://nagayo.sakura.ne.jp/cgi/schedule/schedule.cgi?year=2019&month=10")
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}

		events := parse(doc)

		ctx.Header("Content-Type", "text/calendar")
		ctx.String(http.StatusOK, "BEGIN:VCALENDAR\n"+strings.Join(events, "")+"END:VCALENDAR\n")
	})

	router.Run(":" + port)
}

func parse(doc *goquery.Document) (events []string) {
	var year string
	var month string
	// Year-month header
	innerText := decode(doc.Find("td[colspan='7']").Text())

	// 2019年2月 -> [2019年2月, 2019, 2]
	r := regexp.MustCompile(`^([0-9]{4}).*?([0-9]{1,2}).*$`)
	match := r.FindStringSubmatch(innerText)
	parseLog.Println(match)

	if len(match) != 3 {
		return
	}
	year = match[1]
	month = match[2]

	// nth-of-type(3): "3" represents Tuesday
	doc.Find("tr >td:nth-of-type(3)").Each(func(i int, elem *goquery.Selection) {
		innerText := decode(elem.Text())

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

		event := "BEGIN:VEVENT\n"
		event += "SUMMARY:" + summary + "\n"
		event += "LOCATION:ヤマノミュージックサロン有楽町 〒100-0006\\, 東京都千代田区\\, 有楽町2丁目10番1号\n"
		event += "DTSTART;TZID=Asia/Tokyo:" + fmt.Sprintf("%04s%02s%02sT193000", year, month, date) + "\n"
		event += "DTEND;TZID=Asia/Tokyo:" + fmt.Sprintf("%04s%02s%02sT203000", year, month, date) + "\n"
		event += "END:VEVENT\n"

		events = append(events, event)
	})

	return events
}

func decode(text string) string {
	decoded, _ := ioutil.ReadAll(transform.NewReader(
		strings.NewReader(text),
		japanese.EUCJP.NewDecoder(),
	))

	return string(decoded)
}
