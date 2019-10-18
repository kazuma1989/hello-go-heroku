package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()
	router.LoadHTMLGlob("*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	})

	router.GET("/nagayo.ics", func(c *gin.Context) {
		c.Header("Content-Type", "text/calendar")
		c.String(http.StatusOK, `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:レッスン
LOCATION:ヤマノミュージックサロン有楽町 〒100-0006\, 東京都千代田区\, 有楽町2丁目10番1号
DTSTART:20191018T103000Z
DTEND:20191018T113000Z
END:VEVENT
BEGIN:VEVENT
SUMMARY:レッスン
LOCATION:ヤマノミュージックサロン有楽町 〒100-0006\, 東京都千代田区\, 有楽町2丁目10番1号
DTSTART:20191025T103000Z
DTEND:20191025T113000Z
END:VEVENT
END:VCALENDAR
`)
	})

	router.Run(":" + port)
}
