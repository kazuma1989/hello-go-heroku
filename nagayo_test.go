package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func Example_parseDate_0() {
	doc, _ := goquery.NewDocument(ScheduleURL + "?year=2019&month=10")
	dateCells, _ := parseDate(doc, 0)

	fmt.Println(dateCells)
	// Output: [{1 有楽町山野} {2 有楽町山野} {3 スガナミ多摩} {5 リフラ} {7 ★伊藤} {9 有楽町山野} {11 渋谷} {15 有楽町山野} {17 スガナミ多摩} {18 渋谷} {19 リフラ} {20 ★フォーク有楽町} {21 ★伊藤} {23 有楽町山野} {24 スガナミ多摩} {25 渋谷} {28 ★伊藤} {29 有楽町山野} {30 有楽町山野}]
}

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

func ExampleLocationMap() {
	fmt.Println(LocationMap["有楽町山野"])
	fmt.Println(LocationMap[""])
	// Output: 東京都千代田区有楽町2-10-1 東京交通会館11F
}
