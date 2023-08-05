// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"image"
	"image/png"
	"log"
	"os"

	scraping "github.com/kyoyann/AozoraBunko/Scraping"
	"github.com/oliamb/cutter"
)

func main() {
	if err := scraping.ElementScreenshot("https://www.aozora.gr.jp/cards/000076/files/4996_15646.html", "div.main_text"); err != nil {
		log.Fatalln(err)
	}
	file, _ := os.Open("./elementScreenshot.png")
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	croppedImg, err := cutter.Crop(src, cutter.Config{
		Width:   600,
		Height:  600,
		Options: cutter.Copy,
	})
	if err != nil {
		log.Fatalln(err)
	}
	croppath, _ := os.Create("./cropimage.png")
	png.Encode(croppath, croppedImg)
}
