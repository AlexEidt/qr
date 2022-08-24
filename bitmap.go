package qr

type Bitmap struct {
	width  int
	height int
	data   []uint64
}

func NewBitmap(width, height int) *Bitmap {
	total := width * height
	size := total / 64

	if total%64 != 0 {
		size++
	}

	return &Bitmap{
		width:  width,
		height: height,
		data:   make([]uint64, size),
	}
}

func (b *Bitmap) Width() int {
	return b.width
}

func (b *Bitmap) Height() int {
	return b.height
}

func (b *Bitmap) Copy() *Bitmap {
	new := NewBitmap(b.width, b.height)
	copy(new.data, b.data)
	return new
}

func (b *Bitmap) At(x, y int) bool {
	index := y*b.width + x
	return (b.data[index/64] & (1 << (index & 63))) != 0
}

func (b *Bitmap) Set(x, y int, value bool) {
	index := y*b.width + x
	if value {
		b.data[index/64] |= (1 << (index & 63))
	} else {
		b.data[index/64] &= ^(1 << (index & 63))
	}
}

func (b *Bitmap) Fill(x, y, w, h int, value bool) {
	for i := y; i < y+h; i++ {
		for j := x; j < x+w; j++ {
			b.Set(j, i, value)
		}
	}
}

func (b *Bitmap) Place(x, y int, other *Bitmap) {
	for i := y; i < y+other.Height(); i++ {
		for j := x; j < x+other.Width(); j++ {
			b.Set(j, i, other.At(j-x, i-y))
		}
	}
}

func (b *Bitmap) Invert() {
	for i := 0; i < len(b.data); i++ {
		b.data[i] = ^b.data[i]
	}
}
