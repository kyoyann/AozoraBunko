package scraping

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/kyoyann/AozoraBunko/store"
)

const (
	BASEURL       = "https://www.aozora.gr.jp/"
	FILESIZELIMIT = 7600
	MAINFILEPATH  = "MainTextScreenshot.png"
	MAINSEL       = "div.main_text"
	INFOFILEPATH  = "InformationScreenshot.png"
	INFOSEL       = "div.bibliographical_information"
)

var (
	ErrGetLibraryCardUrl = errors.New("cloud not get librarycard url")
	ErrGetNovelUrl       = errors.New("cloud not get novel url")
	ErrFileSizeOver      = errors.New("file size over")
	ErrGetPage           = errors.New("cloud not get page")
	ErrCopyrightSurvival = errors.New("copyright survival")
)

func GetLibraryCardUrl(charindex, page string) ([]store.Novel, error) {
	ns := make([]store.Novel, 50)
	res, err := http.Get(BASEURL + "index_pages/sakuhin_" + charindex + page)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusOK {
		return nil, ErrGetPage
	}
	defer res.Body.Close()
	fmt.Println(res.Request.URL)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	isFoundUrl := false
	doc.Find("tr[valign='top']").Each(func(i int, s *goquery.Selection) {
		s.Find("td").EachWithBreak(func(j int, ss *goquery.Selection) bool {
			//最初の列だけ数字になっているので、数字に変換してエラーにならなければ最初の列だとわかる
			_, err := strconv.Atoi(ss.Text())
			if err == nil {
				ns[i-1].Title = strings.ReplaceAll(strings.ReplaceAll(ss.Next().Text(), "\n", ""), " ", "")
				ns[i-1].Author = strings.ReplaceAll(ss.Next().Next().Next().Text(), "\n", "")
				ns[i-1].LibraryCardUrl, isFoundUrl = ss.Next().Find("a").Attr("href")
				ns[i-1].LibraryCardUrl = BASEURL + ns[i-1].LibraryCardUrl[3:]
			}
			return false //最初の列だけ確認すれば良い
		})
	})
	if !isFoundUrl {
		return nil, ErrGetLibraryCardUrl
	}
	return ns, nil
}

func GetNovelUrl(url string) (novelUrl string, err error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	isFoundUrl := false
	isSizeOver := false
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	//著作権チェック
	if doc.Find("font[color='red']").Text() != "" {
		return "", ErrCopyrightSurvival
	}
	doc.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		s.Find("td").EachWithBreak(func(j int, ss *goquery.Selection) bool {
			if strings.Contains(ss.Text(), "HTMLファイル") {
				size := ss.Next().Next().Next().Next().Text()
				s, err := strconv.Atoi(size)
				if err != nil || s > FILESIZELIMIT {
					isSizeOver = true
					return false
				}
				novelUrl, isFoundUrl = ss.Next().Next().Find("a").Attr("href")
			}
			return false //最初の列のみ確認すればよい
		})
		return !isFoundUrl
	})
	if isSizeOver {
		return "", ErrFileSizeOver
	}
	if !isFoundUrl {
		return "", ErrGetNovelUrl
	}
	return url[:strings.LastIndex(url, "/")] + novelUrl[1:], nil
}

func Screenshot(urlstr, sel, path string) error {
	// create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// capture screenshot of an element
	var buf []byte
	t := chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Screenshot(sel, &buf, chromedp.NodeVisible),
		chromedp.Emulate(device.IPhone8),
	}
	if err := chromedp.Run(ctx, t); err != nil {
		return err
	}
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		return err
	}
	return nil
}
