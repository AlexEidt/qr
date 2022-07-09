package qr

import (
	"fmt"
	"testing"
)

func assertEquals(actual, expected interface{}) {
	if expected != actual {
		panic(fmt.Sprintf("Expected %v, got %v", expected, actual))
	}
}

func TestRendering(test *testing.T) {
	qr, err := NewQRCode("QR Code", &Options{Version: 31, Error: "H"})
	if err != nil {
		panic(err)
	}
	qr.Render("qr.png", 10)
}

func TestGivenMode(t *testing.T) {
	qr, err := NewQRCode("Hello world +äöpäü+ä 1234", &Options{Mode: Byte})
	if err != nil {
		panic(err)
	}

	assertEquals(qr.mode, Byte)

	// The given string "01234" would normally be represented in numeric mode,
	// however the user specifies they want Alphanumeric mode.
	qr, err = NewQRCode("01234", &Options{Mode: AlphaNum})
	if err != nil {
		panic(err)
	}
	// Make sure that the mode is actually Alphanumeric.
	assertEquals(qr.mode, AlphaNum)
}

func TestErrorLevels(t *testing.T) {
	for _, c := range "LMQH" {
		_, err := NewQRCode("HELLO WORLD 12345 :.", &Options{Error: string(c)})
		if err != nil {
			panic(err)
		}
	}
}

func TestNumericMode(t *testing.T) {
	qr, err := NewQRCode("01234", nil)
	if err != nil {
		panic(err)
	}

	assertEquals(qr.mode, Numeric)
}

func TestAlphaNumMode(t *testing.T) {
	qr, err := NewQRCode("HELLO WORLD 12345 :.", nil)
	if err != nil {
		panic(err)
	}

	assertEquals(qr.mode, AlphaNum)
}

func TestByteMode(t *testing.T) {
	qr, err := NewQRCode("Hello world +äöpäü+ä 1234", nil)
	if err != nil {
		panic(err)
	}

	assertEquals(qr.mode, Byte)
}
