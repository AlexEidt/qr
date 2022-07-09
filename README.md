# `qr`

`qr` is a library for generating QR Codes in pure Go.

```
go get github.com/AlexEidt/qr
```

<p align="center">
  <img width="50%" src="https://github.com/AlexEidt/docs/blob/master/qr/qr.gif">
</p>

## Usage

```go
text := "QR Code"
qrcode, err := qr.NewQRCode(text, &qr.Options{Error: "H"})
if err != nil {
    panic(err)
}

qrcode.Render("qr.png", 10)
```

Below are the functions available for a `QRCode`.

```go
qr.NewQRCode(data string, options *Options) (*QRCode, error)

Version() int
Mode() int // 1: numeric, 2: alphanumeric, 3: byte
ErrorLevel() string // L, M, Q, H
Bitmap() *Bitmap
Render(filename string, scale int) error // .png, .jpg, .svg supported
```

## `Options`

When building a QR Code, certain parameters can be specified such as the Version, Mode and Error Correction Level.

```go
type Options struct {
	Version int
	Mode    int
	Error   string
}
```

Parameter | Description
--- | ---
`Version` | The version of the QR Code to be generated. Must be between 1 and 40. Defaults to lowest version that fits the given data.
`Mode` | The mode of the QR Code to be generated. Must be `qr.Numeric`, `qr.AlphaNum`, or `qr.Byte`. The best fit is found based on the given data. See the **Supported Modes** section below for the characters that can be used in each mode.
`Error` | The error correction level of the QR Code to be generated. Must be `L`, `M`, `Q`, or `H`. Defaults to `L`. Level `L` can correct ~7% of errors, `M` can correct ~15% of errors, `Q` can correct ~25% of errors, and `H` can correct ~30% of errors.

## Supported Modes

Currently, only **Numeric**, **Alphanumeric** and **Binary** modes are supported for data encoding.

Mode | Character Set
--- | ---
**Numeric** | `0123456789`
**Alphanumeric** | `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:`
**Binary** | ISO/IEC 8859-1

## Notes

* Kanji mode is not supported.
* Structured Append mode is not supported.
* Extended Channel Interpretations (ECI) mode is not supported.
* Model 1 QR Codes are not supported.

## Acknowledgements

The following resources were helpful in building this library:

* The `pyqrcode` repository by `mnooner256`: https://github.com/mnooner256/pyqrcode
* The QR Code Tutorial at: http://www.thonky.com/qr-code-tutorial/
* The QR Code Tutorial at: https://www.matchadesign.com/news/blog/qr-code-demystified-part-1/