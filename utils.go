package qr

import (
	"regexp"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func findMode(data string) int {
	digits := regexp.MustCompile("^[0-9]+$")
	if digits.MatchString(data) {
		return Numeric
	}
	alphanum := regexp.MustCompile(`^[0-9A-Z \$\%\*\+\-\.\/\:]+$`)
	if alphanum.MatchString(data) {
		return AlphaNum
	}
	return Byte
}

func length(version, mode int) int {
	var size int
	if version < 10 {
		switch mode {
		case Numeric:
			size = 10
		case AlphaNum:
			size = 9
		case Byte:
			size = 8
		}
	} else if version < 27 {
		switch mode {
		case Numeric:
			size = 12
		case AlphaNum:
			size = 11
		case Byte:
			size = 16
		}
	} else {
		switch mode {
		case Numeric:
			size = 14
		case AlphaNum:
			size = 13
		case Byte:
			size = 16
		}
	}
	return size
}

func maskPattern(mask int) func(int, int) bool {
	switch mask {
	case 0:
		return func(x, y int) bool { return (y+x)%2 == 0 }
	case 1:
		return func(x, y int) bool { return y%2 == 0 }
	case 2:
		return func(x, y int) bool { return x%3 == 0 }
	case 3:
		return func(x, y int) bool { return (y+x)%3 == 0 }
	case 4:
		return func(x, y int) bool { return (y/2+x/3)%2 == 0 }
	case 5:
		return func(x, y int) bool { return (y*x)%2+(y*x)%3 == 0 }
	case 6:
		return func(x, y int) bool { return ((y*x)%2+(y*x)%3)%2 == 0 }
	case 7:
		return func(x, y int) bool { return ((y*x)%3+(y+x)%2)%2 == 0 }
	}
	return nil
}
