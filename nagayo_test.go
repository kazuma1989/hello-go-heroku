package main

import (
	"fmt"
	// テストで使える関数・構造体が用意されているパッケージをimport
	"github.com/PuerkitoBio/goquery"
)

func Example_parseDate_1() {
	doc, _ := goquery.NewDocument(ScheduleURL + "?year=2019&month=10")
	dateCells, _ := parseDate(doc, 1)

	fmt.Println(dateCells)
	// Output: [{20 ★フォーク有楽町}]
}

func Example_parseDate_2() {
	doc, _ := goquery.NewDocument(ScheduleURL + "?year=2019&month=10")
	dateCells, _ := parseDate(doc, 3)

	fmt.Println(dateCells)
	// Output: [{1 有楽町山野} {15 有楽町山野} {29 有楽町山野}]
}
