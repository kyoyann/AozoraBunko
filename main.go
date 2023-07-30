package main

import (
	"fmt"
	"log"

	scraping "github.com/kyoyann/AozoraBunko/Scraping"
)

func main() {
	url, err := scraping.GetNovelUrl("https://www.aozora.gr.jp/cards/002322/card62213.html")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(url)
}
