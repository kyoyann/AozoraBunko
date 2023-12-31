package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/kyoyann/AozoraBunko/scraping"
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
	//作品一覧ページ（https://www.aozora.gr.jp/index_pages/sakuhin_a1.html）から図書カードのURLを取得する。
	for i := 0; i < 45; i++ {
		//索引ごとの作品数は不定なので上限を設定できない
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
