package qr

import (
	"fmt"
	"strings"
)

const (
	Numeric  = 1
	AlphaNum = 2
	Byte     = 4
)

type QRCode struct {
	version    int
	size       int
	mode       int
	errorLevel string
	qr         *Bitmap // The QR Code.
	mask       *Bitmap // The QR Code mask, used to track all functional patterns.
}

type Options struct {
	Version int
	Mode    int
	Error   string
}

func (qr *QRCode) Version() int {
	return qr.version
}

func (qr *QRCode) Mode() int {
	return qr.mode
}

func (qr *QRCode) ErrorLevel() string {
	return qr.errorLevel
}

func (qr *QRCode) Bitmap() *Bitmap {
	return qr.qr.Copy()
}

func NewQRCode(data string, options *Options) (*QRCode, error) {
	qr := &QRCode{}

	if options == nil {
		options = &Options{}
	}

	qr.errorLevel = "L"
	if options.Error != "" {
		if !strings.Contains("LMQH", options.Error) {
			return nil, fmt.Errorf("invalid error level: %s", options.Error)
		}
		qr.errorLevel = options.Error
	}

	qr.version = options.Version
	qr.mode = findMode(data)

	if options.Mode != 0 {
		switch options.Mode {
		case Numeric, AlphaNum, Byte:
			if options.Mode >= qr.mode {
				qr.mode = options.Mode
			} else {
				return nil, fmt.Errorf("could not encode data with given mode")
			}
		default:
			return nil, fmt.Errorf("given mode is not supported")
		}

	}

	optimal := findOptimalVersion(data, qr.mode, strings.Index("LMQH", qr.errorLevel))
	if optimal > 40 {
		return nil, fmt.Errorf("data too large for a QR Code")
	}
	if qr.version == 0 {
		qr.version = optimal
	} else {
		if qr.version < 1 || qr.version > 40 {
			return nil, fmt.Errorf("invalid version number. Must be between 1 and 40")
		}
		if qr.version < optimal {
			return nil, fmt.Errorf("data too large for version %d", qr.version)
		}
	}

	qr.size = qr.version*4 + 17
	qr.qr = NewBitmap(qr.size, qr.size)
	qr.mask = NewBitmap(qr.size, qr.size) // Mask for non-functional area of QR Code.

	buffer := NewBuffer()
	// Add data. First add the mode indicator, then the data length, followed by the data.
	buffer.Add(qr.mode, 4)
	buffer.Add(len(data), length(qr.version, qr.mode))
	encode(buffer, data, qr.mode)

	index := (qr.version-1)*4 + strings.Index("LMQH", qr.errorLevel)

	// Add Termination bits.
	buffer.Add(0, min(4, capacity[index]-buffer.Size()))

	// Add remainder bits to make sure number of bits is a multiple of 8.
	buffer.Add(0, (8-buffer.Size()%8)%8)

	// Add alternating padding bits to fill message to full capacity.
	remaining := (capacity[index] - buffer.Size()) / 8
	for i := 0; i < remaining; i++ {
		if i%2 == 0 {
			buffer.Add(0b11101100, 8)
		} else {
			buffer.Add(0b00010001, 8)
		}
	}

	// "blockData" is either a 3-tuple or a 6-tuple.
	// First value represents number of error correction blocks.
	// Second value represents the total number of codewords.
	// Third value represents the number of data codewords.
	// If "blockData" is a 6-tuple, the next three values represent the
	// same information as the first three values.
	// This means that if there is a 6-tuple, there are multiple error correction
	// blocks with different sizes.
	blockData := blocks[index]

	dbSize := blockData[0]
	if len(blockData) > 3 {
		dbSize += blockData[3]
	}

	bytes := buffer.Bytes()

	dataBlocks := make([][]byte, dbSize)
	// After the data has been encoded into a stream of bytes, the stream must
	// be split into the correct number of blocks as determined by the number
	// of error correction blocks given by "blockData".
	current := 0
	for i := 0; i < dbSize; i++ {
		if i < blockData[0] {
			dataBlocks[i] = bytes[current : current+blockData[2]]
			current += blockData[2]
		} else {
			dataBlocks[i] = bytes[current : current+blockData[5]]
			current += blockData[5]
		}
	}

	errorBlocks := make([][]byte, len(dataBlocks))
	for i, block := range dataBlocks {
		errorBlocks[i] = qr.encodeError(block, i)
	}

	buffer.Clear()
	largestBlock := blockData[2]
	if len(blockData) > 3 {
		largestBlock = max(largestBlock, blockData[5])
	}
	errorwords := blockData[1] - blockData[2]
	// Interleave data blocks:
	// Codeword #1 from block #1, codeword #1 from block #2, ..., codeword #1 from block #n
	// followed by codeword #2 from block #1, codeword #2 from block #2, ..., codeword #2 from block #n
	// ...
	for i := 0; i < largestBlock+errorwords; i++ {
		for _, block := range dataBlocks {
			if i < len(block) {
				buffer.Add(int(block[i]), 8)
			}
		}
	}
	// Interleave error blocks in the same way as data blocks.
	for i := 0; i < errorwords; i++ {
		for _, block := range errorBlocks {
			buffer.Add(int(block[i]), 8)
		}
	}

	// Add functional patterns.
	qr.addPositionPatterns()
	qr.addTimingPatterns()
	qr.addAlignmentPatterns()
	qr.addVersionInformation()

	qr.mask.Invert()

	bitstring := buffer.String()
	mask := qr.findBestMaskPattern(bitstring)
	qr.addFormatInformation(mask)

	qr.placeBits(qr.qr, bitstring, mask)

	// Add Quiet Zone around the QR Code.
	qrcode := NewBitmap(qr.size+8, qr.size+8)
	qrcode.Place(4, 4, qr.qr)
	qr.qr = qrcode

	return qr, nil
}

func (qr *QRCode) addPositionPatterns() {
	// Top Left.
	qr.qr.Fill(0, 0, 7, 7, true)
	qr.qr.Fill(1, 1, 5, 5, false)
	qr.qr.Fill(2, 2, 3, 3, true)
	// Bottom Right.
	qr.qr.Fill(0, qr.size-7, 7, 7, true)
	qr.qr.Fill(1, qr.size-6, 5, 5, false)
	qr.qr.Fill(2, qr.size-5, 3, 3, true)
	// Top Right.
	qr.qr.Fill(qr.size-7, 0, 7, 7, true)
	qr.qr.Fill(qr.size-6, 1, 5, 5, false)
	qr.qr.Fill(qr.size-5, 2, 3, 3, true)

	qr.mask.Fill(0, 0, 9, 9, true)
	qr.mask.Fill(0, qr.size-8, 9, 8, true)
	qr.mask.Fill(qr.size-8, 0, 8, 9, true)
}

func (qr *QRCode) addTimingPatterns() {
	// The dotted line connecting the top left and right and top left and bottom left position patterns.
	for i := 8; i < qr.size-8; i += 2 {
		qr.qr.Set(i, 6, true)
		qr.qr.Set(6, i, true)

		qr.mask.Set(i, 6, true)
		qr.mask.Set(i+1, 6, true)
		qr.mask.Set(6, i, true)
		qr.mask.Set(6, i+1, true)
	}
}

func (qr *QRCode) addAlignmentPatterns() {
	vals := alignmentPositions[qr.version-1]
	for i := 0; i < len(vals); i++ {
		for j := 0; j < len(vals); j++ {
			// Do not include the top left, top right and bottom left alignment patterns since
			// they would block the position patterns.
			if (i == 0 && j == 0) || (i == len(vals)-1 && j == 0) || (i == 0 && j == len(vals)-1) {
				continue
			}
			x, y := vals[i]-2, vals[j]-2

			qr.qr.Fill(x, y, 5, 5, true)
			qr.qr.Fill(x+1, y+1, 3, 3, false)
			qr.qr.Set(x+2, y+2, true)

			qr.mask.Fill(x, y, 5, 5, true)
		}
	}
}

func (qr *QRCode) addVersionInformation() {
	// 3x6 and 6x3 boxes next to Position Detection Patterns in Bottom Left and Top Right corners.
	if qr.version >= 7 {
		qr.mask.Fill(0, qr.size-11, 6, 3, true) // Bottom Left.
		qr.mask.Fill(qr.size-11, 0, 3, 6, true) // Top Right.

		// Fill in version information in QR Code.
		// Bottom Left.
		index := 0
		bits := versionBits[qr.version]
		for x := 0; x < 6; x++ {
			for y := qr.size - 11; y < qr.size-8; y++ {
				qr.qr.Set(x, y, bits&(1<<index) != 0)
				index++
			}
		}
		// Top Right.
		index = 0
		for y := 0; y < 6; y++ {
			for x := qr.size - 11; x < qr.size-8; x++ {
				qr.qr.Set(x, y, bits&(1<<index) != 0)
				index++
			}
		}
	}
}

func (qr *QRCode) addFormatInformation(mask int) {
	err := strings.Index("MLHQ", qr.errorLevel)
	format := formatBits[mask|(err<<3)]

	index := 0
	// Format Information to the right of the top left position pattern.
	for y := 0; y < 9; y++ {
		if y == 6 {
			continue
		}
		qr.qr.Set(8, y, format&(1<<index) != 0)
		index++
	}
	// Format Information to the bottom of the top left position pattern.
	for x := 7; x >= 0; x-- {
		if x == 6 {
			continue
		}
		qr.qr.Set(x, 8, format&(1<<index) != 0)
		index++
	}
	index = 0
	// Format Information to the bottom of the top right position pattern.
	for x := qr.size - 1; x >= qr.size-8; x-- {
		qr.qr.Set(x, 8, format&(1<<index) != 0)
		index++
	}
	// Format Information to the right of the bottom left position pattern.
	for y := qr.size - 7; y < qr.size; y++ {
		qr.qr.Set(8, y, format&(1<<index) != 0)
		index++
	}

	// Dark Module
	qr.qr.Set(8, qr.size-8, true)
}

// Employs the Reed-Solomon Algorithm to generate the error correction codewords
// for a given block of data codewords.
// See: https://www.matchadesign.com/news/blog/qr-code-demystified-part-4/
// for an in-depth explanation.
func (qr *QRCode) encodeError(block []byte, index int) []byte {
	blockData := blocks[(qr.version-1)*4+strings.Index("LMQH", qr.errorLevel)]
	codewords, datawords := blockData[1], blockData[2]
	errorwords := codewords - datawords

	codewordsPerBlock := blockData[2]
	if index >= blockData[0] {
		codewordsPerBlock = blockData[5]
	}

	rserror := make([]byte, len(block)+errorwords)
	copy(rserror, block)

	generator := polynomials[errorwords]

	for i := 0; i < codewordsPerBlock; i++ {
		coefficient := rserror[0]
		rserror = rserror[1:]

		if coefficient == 0 {
			continue
		}

		alphaExp := log[coefficient]

		for g := 0; g < len(generator); g++ {
			val := alphaExp + generator[g]
			if val > 255 {
				val %= 255
			}
			rserror[g] ^= byte(exp[val])
		}
	}

	if len(rserror) < codewordsPerBlock {
		rserror = append(rserror, make([]byte, codewordsPerBlock-len(rserror))...)
	}

	return rserror
}

// Places the data bitstream into the QR Code represented by a bitmap.
func (qr *QRCode) placeBits(bitmap *Bitmap, bitstream string, mask int) {
	inc := -1
	row := qr.size - 1
	index := 0

	mask_func := maskPattern(mask)

	for c := qr.size - 1; c > 0; c -= 2 {
		col := c
		if col <= 6 {
			col -= 1
		}

		for {
			for i := col; i > col-2; i-- {
				if qr.mask.At(i, row) {
					dark := false

					if index < len(bitstream) {
						dark = bitstream[index] == '1'
						index++
					}

					if mask_func(i, row) {
						dark = !dark
					}

					bitmap.Set(i, row, dark)
				}
			}

			row += inc

			if row < 0 || qr.size <= row {
				row -= inc
				inc = -inc
				break
			}
		}
	}
}

func (qr *QRCode) findBestMaskPattern(bitstream string) int {
	bestMask, bestScore := 0, 0

	for mask := 0; mask < 8; mask++ {
		template := qr.qr.Copy()
		qr.placeBits(template, bitstream, mask)
		score := qr.scoreMaskPattern(template)

		if mask == 0 || score < bestScore {
			bestScore = score
			bestMask = mask
		}
	}

	return bestMask
}

func (qr *QRCode) scoreMaskPattern(bitmap *Bitmap) int {
	score := 0

	// Error #1
	// Five consecutive modules of the same color.
	prev := bitmap.At(0, 0)
	for y := 0; y < qr.size; y++ {
		count := 0
		for x := 1; x < qr.size; x++ {
			count = 1
			curr := bitmap.At(x, y)
			for prev == curr && x < qr.size-1 {
				x++
				count++
				curr = bitmap.At(x, y)
			}
			if count >= 5 {
				score += (count - 5) + 3
			}
			prev = curr
		}
		if count >= 5 {
			score += (count - 5) + 3
		}
	}
	prev = bitmap.At(0, 0)
	for x := 0; x < qr.size; x++ {
		count := 0
		for y := 1; y < qr.size; y++ {
			count = 1
			curr := bitmap.At(x, y)
			for prev == curr && y < qr.size-1 {
				y++
				count++
				curr = bitmap.At(x, y)
			}
			if count >= 5 {
				score += (count - 5) + 3
			}
			prev = curr
		}
		if count >= 5 {
			score += (count - 5) + 3
		}
	}

	// Error #2
	// 2x2 blocks of the same color.
	for y := 0; y < qr.size-1; y++ {
		for x := 0; x < qr.size-1; x++ {
			curr := bitmap.At(x, y)
			if curr == bitmap.At(x-1, y) && curr == bitmap.At(x+1, y) && curr == bitmap.At(x+1, y+1) {
				score += 3 // Each 2x2 square adds 3 to the score.
			}
		}
	}

	// Error #3
	// 1011101 Pattern occurences since these conflict with the position patterns.
	for y := 0; y < qr.size-7; y++ {
		for x := 0; x < qr.size-7; x++ {
			pattern := 0b1011101
			checkRows := 0
			checkCols := 0
			for i := 0; i < 7; i++ {
				// Check along rows.
				if bitmap.At(x+i, y) == (pattern&(1<<i) != 0) {
					checkRows++
				}
				// Check along columns.
				if bitmap.At(x, y+i) == (pattern&(1<<i) != 0) {
					checkCols++
				}
			}
			if checkRows == 7 {
				score += 40 // Each 1011101 pattern adds 40 to the score.
			}
			if checkCols == 7 {
				score += 40
			}
		}
	}

	// Error #4
	// Deviation from being 50% dark modules.
	darkCount := 0
	for y := 0; y < qr.size; y++ {
		for x := 0; x < qr.size; x++ {
			if bitmap.At(x, y) {
				darkCount++
			}
		}
	}

	ratio := float64(darkCount) / float64(qr.size*qr.size)
	ratio = ratio*100 - 50
	score += int((float64(abs(int(ratio))) / 5) * 10)

	return score
}
