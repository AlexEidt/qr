package qr

import (
	"regexp"
	"strconv"
	"strings"
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

func findOptimalVersion(data string, mode, errorIndex int) int {
	buffer := NewBuffer()
	encode(buffer, data, mode)

	for version := 1; version <= 40; version++ {
		index := (version-1)*4 + errorIndex
		blockData := blocks[index]
		maxbytes := blockData[0] * blockData[2]
		if len(blockData) > 3 {
			maxbytes += blockData[3] * blockData[5]
		}

		size := 4 + length(version, mode) + buffer.Size()
		size += max(min(4, capacity[index]-size), 0)
		size += (8 - size%8) % 8

		if size/8 <= maxbytes {
			return version
		}
	}

	return 41
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

func encode(buffer *Buffer, data string, mode int) {
	switch mode {
	case Numeric:
		for i := 0; i < len(data); i += 3 {
			str := string(data[i])
			j := i + 1
			for j < len(data) && j < i+3 {
				str += string(data[j])
				j++
			}
			var dlen int
			switch len(str) {
			case 1:
				dlen = 4
			case 2:
				dlen = 7
			case 3:
				dlen = 10
			}
			n, _ := strconv.Atoi(str)
			buffer.Add(n, dlen)
		}
	case AlphaNum:
		chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"
		for i := 0; i < len(data); i += 2 {
			str := string(data[i])
			if i+1 < len(data) {
				str += string(data[i+1])
			}
			if len(str) > 1 {
				val := strings.Index(chars, string(str[0])) * 45
				val += strings.Index(chars, string(str[1]))
				buffer.Add(val, 11)
			} else {
				buffer.Add(strings.Index(chars, string(str[0])), 6)
			}
		}
	case Byte:
		for _, b := range []byte(data) {
			buffer.Add(int(b), 8)
		}
	}
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
