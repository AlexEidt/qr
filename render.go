package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func (qr *QRCode) Render(filename string, scale int) error {
	extension := filepath.Ext(filename)
	switch extension {
	case ".png", ".jpg", ".jpeg":
		return qr.renderRaster(filename, scale)
	case ".svg":
		return qr.renderVector(filename, scale)
	}
	return fmt.Errorf("unsupported file extension: %s", extension)
}

func (qr *QRCode) renderVector(filename string, scale int) error {
	writer := NewBuffer()

	template := `<svg version="1.1" xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`
	writer.Write(fmt.Sprintf(template, qr.qr.Width()*scale, qr.qr.Height()*scale))

	writer.Write(`<rect width="100%" height="100%" fill="white" />`)

	for h := 0; h < qr.qr.Height(); h++ {
		for w := 0; w < qr.qr.Width(); w++ {
			if qr.qr.Get(w, h) {
				template := `<rect x="%d" y="%d" width="%d" height="%d" fill="#000" />`
				writer.Write(fmt.Sprintf(template, w*scale, h*scale, scale, scale))
			}
		}
	}

	writer.Write("</svg>")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(writer.String()))
	if err != nil {
		return err
	}

	return nil
}

func (qr *QRCode) renderRaster(filename string, scale int) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if scale < 1 {
		scale = 1
	}

	image := image.NewRGBA(image.Rect(0, 0, qr.qr.Width()*scale, qr.qr.Height()*scale))
	for h := 0; h < qr.qr.Height(); h++ {
		for w := 0; w < qr.qr.Width(); w++ {
			var c color.RGBA
			if qr.qr.Get(w, h) {
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
