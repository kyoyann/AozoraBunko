package image

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/oliamb/cutter"
)

const (
	WHITE uint32 = 65535
	WIDTH int    = 640
)

func CreatePostImages(path string) (int, error) {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	mh := bounds.Max.Y - 1
	sh := 0
	cn := 1
	//800ピクセルごとに切り出す
	for eh := 800; ; eh += 800 {
		//最後はmaxhを切り出して保存する
		if eh >= mh {
			//切り出し範囲が全て白色の場合は画像を生成しない
			if isAllWhite(img, sh, mh) {
				break
			}
			if err := createCropImage(img, sh, mh, cn); err != nil {
				return 0, err
			}
			break
		}
		//切り出す部分が文字にならないように調整する。
		eh = getWiteLine(img, eh)
		//切り出し範囲が全て白色の場合は画像を生成しない
		if isAllWhite(img, sh, eh) {
			sh = eh
			continue
		}
		if err := createCropImage(img, sh, eh, cn); err != nil {
			return 0, err
		}
		cn++
		sh = eh
		//最後まで切り取ったら終了
		if eh == mh {
			break
		}
	}
	return cn, nil
}

//TODO　全て白色の画像は保存しないようにする。IsAllWhite関数作って最初にチェックする。
func createCropImage(img image.Image, starth, endh, index int) error {
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:   WIDTH,
		Height:  endh - starth,
		Anchor:  image.Point{0, starth},
		Options: cutter.Copy,
	})
	if err != nil {
		return err
	}

	croppath, err := os.Create(fmt.Sprintf("./cropimage_%d.png", index))
	if err != nil {
		return err
	}

	err = png.Encode(croppath, croppedImg)
	if err != nil {
		return err
	}
	return nil
}

func isWhitePoint(img image.Image, w, h int) bool {
	r, g, b, a := img.At(w, h).RGBA()
	if r == WHITE && g == WHITE && b == WHITE && a == WHITE {
		return true
	}
	return false
}

//列が全て白色（＝文字が含まれていない）か判定する
func isWhiteLine(img image.Image, checkline int) bool {
	for w := 0; w < WIDTH; w++ {
		if !isWhitePoint(img, w, checkline) {
			return false
		}
	}
	return true
}

//特定の範囲が全て白色（＝文字が含まれていない）か判定する
func isAllWhite(img image.Image, starth, endh int) bool {
	for w := 0; w < WIDTH; w++ {
		for h := starth; h <= endh; h++ {
			if !isWhitePoint(img, w, h) {
				return false
			}
		}
	}
	return true
}

func getWiteLine(img image.Image, h int) int {
	for ; ; h++ {
		if isWhiteLine(img, h) {
			return h
		}
	}
}
