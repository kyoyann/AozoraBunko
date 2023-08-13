// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	image "github.com/kyoyann/AozoraBunko/Image"
	scraping "github.com/kyoyann/AozoraBunko/Scraping"
	twitter "github.com/kyoyann/AozoraBunko/Twitter"
	"github.com/kyoyann/AozoraBunko/store"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//エラーのファイル名と行数を表示
	log.SetFlags(log.Lshortfile)

	db, err := sql.Open("sqlite3", "./store/aozora.db")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	ns, err := store.GetNotPostedNovels(db)
	if err != nil {
		log.Fatalln(err)
	}
	var url string
	var index int
Loop:
	for i, n := range ns {
		fmt.Println(n)
		url, err = scraping.GetNovelUrl(n.LibraryCardUrl)
		switch err {
		case scraping.ErrFileSizeOver:
			if err := store.UpdatePostStatus(db, n.ID, store.FileSizeOver); err != nil {
				log.Fatalln(err)
			}
		case scraping.ErrCopyrightSurvival:
			if err := store.UpdatePostStatus(db, n.ID, store.CopyrightSurvival); err != nil {
				log.Fatalln(err)
			}
		case scraping.ErrGetNovelUrl:
			if err := store.UpdatePostStatus(db, n.ID, store.CloudNotGetNovelUrl); err != nil {
				log.Fatalln(err)
			}
		case nil:
			if err := store.UpdatePostStatus(db, n.ID, store.Posted); err != nil {
				log.Fatalln(err)
			}
			index = i
			//投稿可能な小説を取得できたらfor文を抜ける
			break Loop
		default:
			log.Fatalln(err)
		}
		//連続してリクエストを送らない
		time.Sleep(time.Second * 3)
	}

	if err := scraping.Screenshot(url, scraping.MAINSEL, scraping.MAINFILEPATH); err != nil {
		log.Fatalln(err)
	}

	if err := scraping.Screenshot(url, scraping.INFOSEL, scraping.INFOFILEPATH); err != nil {
		log.Fatalln(err)
	}

	cn, err := image.CreatePostImages(scraping.MAINFILEPATH)
	if err != nil {
		log.Fatalln(err)
	}

	//ファイルサイズを制限しているため、画像が3枚以上になることはないが念の為エラー処理を入れておく
	if cn >= 4 {
		log.Fatalln("too many images")
	}

	var ids []string
	for i := 1; i <= cn; i++ {
		id, err := twitter.PostImage(fmt.Sprintf("./cropimage_%d.png", i))
		if err != nil {
			log.Fatalln(err)
		}
		ids = append(ids, id)
	}

	id, err := twitter.PostImage(scraping.INFOFILEPATH)
	if err != nil {
		log.Fatalln(err)
	}
	ids = append(ids, id)

	if err := twitter.PostTweet(ns[index].Title, ns[index].Author, ids); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("post", ns[index].Title, ns[index].Author)
}
