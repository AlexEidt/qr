package qr

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func (qr *QRCode) Render(filename string, scale int) error {
	qrcode := NewBitmap(qr.size+8, qr.size+8)
	qrcode.Place(4, 4, qr.qr)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if scale < 1 {
		scale = 1
	}

	image := image.NewRGBA(image.Rect(0, 0, qrcode.Width()*scale, qrcode.Height()*scale))
	for h := 0; h < qrcode.Height(); h++ {
		for w := 0; w < qrcode.Width(); w++ {
			var c color.RGBA
			if qrcode.Get(w, h) {
				c = color.RGBA{0, 0, 0, 255}
			} else {
				c = color.RGBA{255, 255, 255, 255}
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					image.Set(w*scale+dx, h*scale+dy, c)
				}
			}
		}
	}

	switch filepath.Ext(filename) {
	case ".png":
		if err := png.Encode(f, image); err != nil {
			return err
		}
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(f, image, nil); err != nil {
			return err
		}
	}

	return nil
}
