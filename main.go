package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/kyoyann/AozoraBunko/image"
	"github.com/kyoyann/AozoraBunko/scraping"
	"github.com/kyoyann/AozoraBunko/store"
	"github.com/kyoyann/AozoraBunko/twitter"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
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
		//連続してリクエストを送らない
		time.Sleep(time.Second * 3)
		switch err {
		case scraping.ErrFileSizeOver:
			if err := store.UpdatePostStatus(db, n.ID, store.FILESIZEOVER); err != nil {
				log.Println(err)
			}
		case scraping.ErrCopyrightSurvival:
			if err := store.UpdatePostStatus(db, n.ID, store.COPYRIGHT_SURVIVAL); err != nil {
				log.Println(err)
			}
		case scraping.ErrGetNovelUrl:
			if err := store.UpdatePostStatus(db, n.ID, store.NOT_GET_NOVELURL); err != nil {
				log.Println(err)
			}
		case nil:
			index = i
			//投稿可能な小説を取得できたらfor文を抜ける
			break Loop
		default:
			log.Fatalln(err)
		}
	}

	if err := scraping.Screenshot(url); err != nil {
		image.DeleteImages()
		log.Fatalln(err)
	}

	cn, err := image.CreatePostImages(scraping.MAINFILEPATH)
	if err != nil {
		image.DeleteImages()
		log.Fatalln(err)
	}

	//ファイルサイズを制限しているため、画像が4枚以上になることはないが念の為エラー処理を入れておく
	if cn >= 4 {
		image.DeleteImages()
		log.Fatalln("too many images")
	}

	var ids []string
	for i := 1; i <= cn; i++ {
		id, err := twitter.PostImage(fmt.Sprintf("./cropimage_%d.png", i))
		if err != nil {
			image.DeleteImages()
			log.Fatalln(err)
		}
		ids = append(ids, id)
	}

	id, err := twitter.PostImage(scraping.INFOFILEPATH)
	if err != nil {
		image.DeleteImages()
		log.Fatalln(err)
	}
	ids = append(ids, id)

	if err := twitter.PostTweet(ns[index].Title, ns[index].Author, ids); err != nil {
		image.DeleteImages()
		log.Fatalln(err)
	}
	fmt.Println("post", ns[index].Title, ns[index].Author)

	if err := store.UpdatePostStatus(db, ns[index].ID, store.POSTED); err != nil {
		image.DeleteImages()
		log.Fatalln(err)
	}
	image.DeleteImages()
}
