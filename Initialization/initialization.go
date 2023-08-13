package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	scraping "github.com/kyoyann/AozoraBunko/Scraping"
	"github.com/kyoyann/AozoraBunko/store"
	_ "github.com/mattn/go-sqlite3"
)

//array,sliceは定数として定義できない
var charindexs = []string{
	"a", "i", "u", "e", "o",
	"ka", "ki", "ku", "ke", "ko",
	"sa", "si", "su", "se", "so",
	"ta", "ti", "tu", "te", "to",
	"na", "ni", "nu", "ne", "no",
	"ha", "hi", "hu", "he", "ho",
	"ma", "mi", "mu", "me", "mo",
	"ya", "yu", "yo",
	"ra", "ri", "ru", "re", "ro",
	"wa", "zz",
}

func main() {
	db, err := sql.Open("sqlite3", "../store/aozora.db")
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < 45; i++ {
		for j := 1; ; j++ {
			ns, err := scraping.GetLibraryCardUrl(charindexs[i], strconv.Itoa(j))
			if errors.Is(err, scraping.ErrGetPage) {
				//ページがなくなったらBreak
				break
			}
			if err != nil {
				log.Fatalln(err)
			}
			for _, vv := range ns {
				if vv.Title == "" {
					//最後のページは50件未満の場合があるため、Titleが空白ならそれ以降もberakしてInsetしないようにする
					break
				} else {
					if err := store.Insert(db, vv); err != nil {
						log.Fatalln(err)
					}
				}
			}
			//連続してリクエストを送信しない
			time.Sleep(3 * time.Second)
		}
	}

}

/*func Insert(n scraping.Novel) error {
	// データベースのコネクションを開く
	db, err := sql.Open("sqlite3", "./novel.db")
	if err != nil {
		return err
	}

	// テーブル作成
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS "NOVEL" ("ID" INTEGER PRIMARY KEY AUTOINCREMENT, "TITLE" VARCHAR(255), "AUTHOR" VARCHAR(255), "LIBRARYCARDURL" VARCHAR(255), "DELIVERYED" INTEGER)`,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO NOVEL(TITLE,AUTHOR,LIBRARYCARDURL,DELIVERYED) VALUES(?,?,?,?)`,
		n.Title,
		n.Author,
		n.LibraryCardUrl,
		0,
	)
	if err != nil {
		return err
	}
	return nil
}*/
