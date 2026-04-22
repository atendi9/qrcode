package encoder

import (
	"strings"
	"bytes"
	"fmt"
	"log"
)

// Bitset represents a set of bits backed by a byte array.
type Bitset struct {
	// numBits is the number of bits currently stored in the set.
	numBits int
	// bits is the underlying byte array holding the bit values.
	bits []byte
}

// New creates and returns a new Bitset initialized with the provided bits.
func New(v ...Bit) *Bitset {
	b := &Bitset{numBits: 0, bits: make([]byte, 0)}
	b.AppendBits(v...)

	return b
}

// NewFromBase2String creates a new Bitset from a base-2 string representation (e.g., "1010").
// It ignores space characters and panics if an invalid character is encountered.
func NewFromBase2String(b2string string) *Bitset {
	b := &Bitset{numBits: 0, bits: make([]byte, 0)}

	for _, c := range b2string {
		switch c {
		case Black.Representation():
			b.AppendBits(Black)
		case White.Representation():
			b.AppendBits(White)
		case ' ':
		default:
			log.Panicf("Invalid char %c in NewFromBase2String", c)
		}
	}

	return b
}

// Clone creates and returns a deep copy of the provided Bitset.
func Clone(from *Bitset) *Bitset {
	return &Bitset{numBits: from.numBits, bits: from.bits[:]}
}

// Append appends the bits from another Bitset to the current one.
func (b *Bitset) Append(other *Bitset) {
	b.ensureCapacity(other.numBits)

	for i := range other.numBits {
		if other.At(i) {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

// AppendBits appends a sequence of Bit values to the Bitset.
func (b *Bitset) AppendBits(bits ...Bit) {
	b.ensureCapacity(len(bits))

	for _, v := range bits {
		if v {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

// AppendBools appends a sequence of boolean values to the Bitset.
func (b *Bitset) AppendBools(bits ...bool) {
	b.ensureCapacity(len(bits))

	for _, v := range bits {
		if v {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}
		b.numBits++
	}
}

// AppendNumBools appends a specific boolean value a given number of times to the Bitset.
func (b *Bitset) AppendNumBools(num int, value bool) {
	for range num {
		b.AppendBools(value)
	}
}

// AppendBytes appends all the bits from the provided byte slice to the Bitset.
func (b *Bitset) AppendBytes(data []byte) {
	for _, d := range data {
		b.AppendByte(d, 8)
	}
}

// AppendByte appends a specified number of bits from a byte to the Bitset.
// It panics if numBits is greater than 8.
func (b *Bitset) AppendByte(value byte, numBits int) {
	b.ensureCapacity(numBits)

	if numBits > 8 {
		log.Panicf("numBits %d out of range 0-8", numBits)
	}

	for i := numBits - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}

		b.numBits++
	}
}

// AppendUint32 appends a specified number of bits from a uint32 value to the Bitset.
// It panics if numBits is greater than 32.
func (b *Bitset) AppendUint32(value uint32, numBits int) {
	b.ensureCapacity(numBits)

	if numBits > 32 {
		log.Panicf("numBits %d out of range 0-32", numBits)
	}

	for i := numBits - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			b.bits[b.numBits/8] |= 0x80 >> uint(b.numBits%8)
		}

		b.numBits++
	}
}

// Bit represents a single bit mapped to a boolean value.
type Bit bool

// Bits converts a slice of booleans into a slice of Bit values.
func Bits(b ...bool) []Bit {
	result := make([]Bit, len(b))
	for i, v := range b {
		result[i] = Bit(v)
	}
	return result
}

// Representation returns the rune representation of the bit ('1' for true, '0' for false).
func (b Bit) Representation() rune {
	if b {
		return '1'
	}
	return '0'
}

// Bool returns the underlying boolean value of the Bit.
func (b Bit) Bool() bool {
	return bool(b)
}

const (
	// White represents an unset bit (false).
	White Bit = false
	// Black represents a set bit (true).
	Black Bit = true
)

// Equals returns true if the current Bitset and the other Bitset are equal in length and content.
func (b *Bitset) Equals(other *Bitset) bool {
	if b.numBits != other.numBits {
		return White.Bool()
	}

	if !bytes.Equal(b.bits[0:b.numBits/8], other.bits[0:b.numBits/8]) {
		return White.Bool()
	}

	for i := 8 * (b.numBits / 8); i < b.numBits; i++ {
		a := (b.bits[i/8] & (0x80 >> byte(i%8)))
		b2 := (other.bits[i/8] & (0x80 >> byte(i%8)))

		if a != b2 {
			return White.Bool()
		}
	}

	return Black.Bool()
}

// Bits returns all the bits in the Bitset as a slice of Bit values.
func (b *Bitset) Bits() []Bit {
	result := make([]Bit, b.numBits)

	var i int
	for i = range b.numBits {
		result[i] = (b.bits[i/8] & (0x80 >> byte(i%8))) != 0
	}

	return result
}

// At returns the boolean value of the bit at the specified index.
// It panics if the index is out of bounds.
func (b *Bitset) At(index int) bool {
	if index >= b.numBits {
		log.Panicf("Index %d out of range", index)
	}

	return (b.bits[index/8] & (0x80 >> byte(index%8))) != 0
}

// ByteAt constructs and returns a byte from up to 8 bits starting at the specified index.
// It panics if the index is out of bounds.
func (b *Bitset) ByteAt(index int) byte {
	if index < 0 || index >= b.numBits {
		log.Panicf("Index %d out of range", index)
	}

	var result byte

	for i := index; i < index+8 && i < b.numBits; i++ {
		result <<= 1
		if b.At(i) {
			result |= 1
		}
	}

	return result
}

// ensureCapacity ensures that the underlying byte array has enough capacity
// to accommodate the specified number of additional bits.
func (b *Bitset) ensureCapacity(numBits int) {
	numBits += b.numBits

	newNumBytes := numBits / 8
	if numBits%8 != 0 {
		newNumBytes++
	}

	if len(b.bits) >= newNumBytes {
		return
	}

	b.bits = append(b.bits, make([]byte, newNumBytes+2*len(b.bits))...)
}

// Len returns the current number of bits in the Bitset.
func (b *Bitset) Len() int {
	return b.numBits
}

// String returns a string representation of the Bitset, including its length and bit sequence.
func (b *Bitset) String() string {
	var bitString strings.Builder
	for i := range b.numBits {
		if (i % 8) == 0 {
			bitString .WriteString(" ")
		}

		if (b.bits[i/8] & (0x80 >> byte(i%8))) != 0 {
			bitString .WriteString(string(Black.Representation()))
		} else {
			bitString .WriteString(string(White.Representation()))
		}
	}

	return fmt.Sprintf("numBits=%d, bits=%s", b.numBits, bitString.String())
}

// Substr creates and returns a new Bitset containing the bits from the start index up to the end index.
// It panics if the indices are out of bounds or if start is greater than end.
func (b *Bitset) Substr(start int, end int) *Bitset {
	if start > end || end > b.numBits {
		log.Panicf("Out of range start=%d end=%d numBits=%d", start, end, b.numBits)
	}

	result := New()
	result.ensureCapacity(end - start)

	for i := start; i < end; i++ {
		if b.At(i) {
			result.bits[result.numBits/8] |= 0x80 >> uint(result.numBits%8)
		}
		result.numBits++
	}

	return result
}
