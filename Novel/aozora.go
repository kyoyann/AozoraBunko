package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Novel struct {
	FileType    string
	Compression string
	FileLink    string
	SignFormat  string
}

const (
	ZIPFILE  = "novel.zip"
	TEXTFILE = "novel.txt"
)

func main() {
	n, err := GetTextFileInfo("https://www.aozora.gr.jp/cards/000081/card45630.html")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(n)
	err = DownloadTextFile("https://www.aozora.gr.jp/cards/000081" + n.FileLink)
	if err != nil {
		log.Fatalln(err)
	}

	err = Unzip()
	if err != nil {
		log.Fatalln(err)
	}

	err = ReadText()
	if err != nil {
		log.Fatalln(err)
	}

}

func GetTextFileInfo(url string) (*Novel, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	n := Novel{}
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td").Each(func(j int, ss *goquery.Selection) {
			if strings.Contains(ss.Text(), "テキストファイル") {
				n.FileType = strings.Replace(strings.Replace(ss.Text(), " ", "", -1), "\n", "", -1)
				n.Compression = strings.Replace(strings.Replace(ss.Next().Text(), " ", "", -1), "\n", "", -1)
				n.FileLink, _ = ss.Next().Next().Find("a").Attr("href")
				n.FileLink = n.FileLink[1:]
				n.SignFormat = strings.Replace(strings.Replace(ss.Next().Next().Next().Text(), " ", "", -1), "\n", "", -1)
			}
		})
	})
	return &n, nil
}

func DownloadTextFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(ZIPFILE)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err

}
func Unzip() error {
	r, err := zip.OpenReader(ZIPFILE)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := TEXTFILE
		if f.FileInfo().IsDir() {
			continue
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadText() error {
	sjisFile, err := os.Open(TEXTFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer sjisFile.Close()
	reader := transform.NewReader(sjisFile, japanese.ShiftJIS.NewDecoder())
	utf8File, err := os.Create("./utf-8.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer utf8File.Close()

	// 書き込み
	tee := io.TeeReader(reader, utf8File)
	s := bufio.NewScanner(tee)
	for s.Scan() {
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")

	bytes, err := ioutil.ReadFile(TEXTFILE)
	if err != nil {
		return err
	}

	str := string(bytes)
	fmt.Println(utf8.RuneCountInString(str))
	//fmt.Println(str)

	return nil
}