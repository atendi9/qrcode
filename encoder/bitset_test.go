package encoder

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/capivara/langs"
)

func TestNew(t *testing.T) {
	a := assert.New(langs.EN_US, t)
	bs := New(Black, White, Black)

	assert.Equal(a, bs.Len(), 3)
	assert.Equal(a, bs.At(0), true)
	assert.Equal(a, bs.At(1), false)
	assert.Equal(a, bs.At(2), true)
}

func TestNewFromBase2String(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	t.Run("ValidBinaryStringWithSpaces", func(t *testing.T) {
		bs := NewFromBase2String("101 0")
		assert.Equal(a, bs.Len(), 4)
		assert.Equal(a, bs.At(0), true)
		assert.Equal(a, bs.At(1), false)
		assert.Equal(a, bs.At(2), true)
		assert.Equal(a, bs.At(3), false)
	})
}

func TestAppendMethods(t *testing.T) {
	a := assert.New(langs.EN_US, t)
	bs := New()

	t.Run("AppendBits", func(t *testing.T) {
		bs.AppendBits(Black)
		assert.Equal(a, bs.Len(), 1)
	})

	t.Run("AppendBools", func(t *testing.T) {
		bs.AppendBools(false, true)
		assert.Equal(a, bs.Len(), 3)
		assert.Equal(a, bs.At(1), false)
		assert.Equal(a, bs.At(2), true)
	})

	t.Run("AppendNumBools", func(t *testing.T) {
		bs.AppendNumBools(2, false)
		assert.Equal(a, bs.Len(), 5)
		assert.Equal(a, bs.At(4), false)
	})

	t.Run("AppendBitset", func(t *testing.T) {
		other := New(Black, Black)
		bs.Append(other)
		assert.Equal(a, bs.Len(), 7)
		assert.Equal(a, bs.At(5), true)
		assert.Equal(a, bs.At(6), true)
	})
}

func TestNumericAppending(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	t.Run("AppendByte", func(t *testing.T) {
		bs := New()
		bs.AppendByte(5, 3)
		assert.Equal(a, bs.Len(), 3)
		assert.Equal(a, bs.At(0), true)
		assert.Equal(a, bs.At(1), false)
		assert.Equal(a, bs.At(2), true)
	})

	t.Run("AppendUint32", func(t *testing.T) {
		bs2 := New()
		bs2.AppendUint32(10, 4)
		assert.Equal(a, bs2.Len(), 4)
		assert.Equal(a, bs2.At(0), true)
		assert.Equal(a, bs2.At(1), false)
		assert.Equal(a, bs2.At(2), true)
		assert.Equal(a, bs2.At(3), false)
	})

	t.Run("AppendBytes", func(t *testing.T) {
		bs3 := New()
		bs3.AppendBytes([]byte{0x80}) // 10000000
		assert.Equal(a, bs3.Len(), 8)
		assert.Equal(a, bs3.At(0), true)
		assert.Equal(a, bs3.At(1), false)
	})
}

func TestAccessorsAndConversions(t *testing.T) {
	a := assert.New(langs.EN_US, t)
	bs := NewFromBase2String("11110000 1")

	assert.Equal(a, bs.Len(), 9)
	assert.Equal(a, bs.ByteAt(0), byte(0xF0))
	assert.Equal(a, bs.ByteAt(1), byte(0xE1)) 

	bits := bs.Bits()
	assert.Equal(a, len(bits), 9)
	assert.Equal(a, bits[0].Bool(), true)

	t.Run("StringRepresentation", func(t *testing.T) {
		expectedStr := "numBits=9, bits= 11110000 1"
		assert.Equal(a, bs.String(), expectedStr)
	})
}

func TestCloneAndEquals(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	bs1 := NewFromBase2String("1101")
	bs2 := Clone(bs1)

	assert.Equal(a, bs1.Equals(bs2), true)

	bs2.AppendBits(White)
	assert.Equal(a, bs1.Equals(bs2), false)

	bs3 := NewFromBase2String("1100")
	assert.Equal(a, bs1.Equals(bs3), false)
}

func TestSubstr(t *testing.T) {
	a := assert.New(langs.EN_US, t)

	bs := NewFromBase2String("101101")
	sub := bs.Substr(1, 4)

	assert.Equal(a, sub.Len(), 3)
	assert.Equal(a, sub.At(0), false)
	assert.Equal(a, sub.At(1), true)
	assert.Equal(a, sub.At(2), true)
}

func TestCapacityManagement(t *testing.T) {
	a := assert.New(langs.EN_US, t)
	bs := New()

	for range 100 {
		bs.AppendBits(Black)
	}

	assert.Equal(a, bs.Len(), 100)
	assert.Equal(a, bs.At(99), true)
}
