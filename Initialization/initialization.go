package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

type Novel struct {
	Title     string
	Author    string
	AozoraURL string
	BookURL   string
}

func main() {
	novels := make([]Novel, 4700) //4606
	for i := 1; i <= 93; i++ {    //93
		GetNovelInfor(novels, i)
		time.Sleep(11 * time.Second)
	}
	for i := 0; i < 4606; i++ {
		GetNovelURL(novels, novels[i].AozoraURL, i)
		time.Sleep(11 * time.Second)
	}

	for i := 0; i < 4606; i++ {
		Insert(novels[i], i)
	}

	/*for i := 0; i < 4606; i++ {
		if novels[i].Title != "" {
			fmt.Printf("%d,%s,%s,%s,%s\n", i, novels[i].Title, novels[i].Author, novels[i].AozoraURL, novels[i].BookURL)
		}
	}*/
}

func GetNovelInfor(novels []Novel, page int) {
	url := fmt.Sprintf("https://bungo-search.com/authors/all/categories/flash/books?page=%d", page)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	/* 見出しの取得 */
	fmt.Println("get novel infor Page=", page)
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	doc.Find("[class='flex items-center text-sm mb-2']").Each(func(i int, s *goquery.Selection) {
		novel := Novel{}
		s.Find(".text-link").Each(func(j int, ss *goquery.Selection) {
			if j%2 == 0 {
				novel.Title = strings.ReplaceAll(ss.Text(), "\n", "")
				novel.AozoraURL, _ = ss.Attr("href")
				novel.AozoraURL = "https://bungo-search.com/" + novel.AozoraURL
			} else {
				novel.Author = strings.ReplaceAll(ss.Text(), "\n", "")
				novels[(page-1)*50+i] = novel
			}
		})
	})
}

func GetNovelURL(novels []Novel, url string, index int) {
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	fmt.Println("get novel url index=", index)
	/* 見出しの取得 */
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	novels[index].BookURL, _ = doc.Find("[class='inline-block w-full sm:w-auto rounded border border-blue-500 text-blue-500 px-4 py-2 hover:bg-blue-100']").Attr("href")
}

func Insert(novel Novel, id int) {
	// データベースのコネクションを開く
	db, err := sql.Open("sqlite3", "./novel.db")
	if err != nil {
		panic(err)
	}

	// テーブル作成
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS "NOVEL" ("ID" INTEGER PRIMARY KEY AUTOINCREMENT, "TITLE" VARCHAR(255), "AUTHOR" VARCHAR(255), "URL" VARCHAR(255))`,
	)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(
		`UPDATE NOVEL SET URL = ? WHERE id = ?`,
		novel.BookURL,
		id+1,
	)
	if err != nil {
		panic(err)
	}

}
