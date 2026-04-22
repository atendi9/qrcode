// Package encoder provides functionality for generating, formatting, and exporting QR codes.
// It handles data encoding, error correction, and rendering the final QR code into various formats
// such as images, bit matrices, and terminal strings.
package encoder

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
)

// QRCode represents a generated QR code with its content, configuration, and structural data.
type QRCode struct {
	Content         string
	Level           RecoveryLevel
	VersionNumber   int
	ForegroundColor color.Color
	BackgroundColor color.Color
	DisableBorder   bool
	encoder         *dataEncoder
	version         qrCodeVersion
	data            *Bitset
	symbol          *symbol
	mask            int
}

// NewQRCode creates a new QRCode with the given content and error recovery level.
// It automatically determines the smallest possible QR code version that can fit the content.
func NewQRCode(content string, level RecoveryLevel) (*QRCode, error) {
	encoders := []dataEncoderType{
		dataEncoderType1To9,
		dataEncoderType10To26,
		dataEncoderType27To40,
	}

	var (
		encoder       *dataEncoder
		encoded       *Bitset
		chosenVersion *qrCodeVersion
		err           error
	)

	for _, t := range encoders {
		encoder = newDataEncoder(t)
		encoded, err = encoder.encode([]byte(content))
		if err != nil {
			continue
		}
		chosenVersion = chooseQRCodeVersion(level, encoder, encoded.Len())

		if chosenVersion != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	if chosenVersion == nil {
		return nil, errors.New("content too long to encode")
	}

	q := &QRCode{
		Content:         content,
		Level:           level,
		VersionNumber:   chosenVersion.version,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
		encoder:         encoder,
		data:            encoded,
		version:         *chosenVersion,
	}

	return q, nil
}

// NewWithForcedVersion creates a new QRCode using a specific version and error recovery level.
// It returns an error if the content exceeds the capacity of the specified version.
func NewWithForcedVersion(content string, version int, level RecoveryLevel) (*QRCode, error) {
	var encoder *dataEncoder

	switch {
	case version >= 1 && version <= 9:
		encoder = newDataEncoder(dataEncoderType1To9)
	case version >= 10 && version <= 26:
		encoder = newDataEncoder(dataEncoderType10To26)
	case version >= 27 && version <= 40:
		encoder = newDataEncoder(dataEncoderType27To40)
	default:
		return nil, fmt.Errorf("invalid version %d (expected 1-40 inclusive)", version)
	}

	var encoded *Bitset
	encoded, err := encoder.encode([]byte(content))

	if err != nil {
		return nil, err
	}

	chosenVersion := getQRCodeVersion(level, version)

	if chosenVersion == nil {
		return nil, errors.New("cannot find QR Code version")
	}

	if encoded.Len() > chosenVersion.numDataBits() {
		return nil, fmt.Errorf("cannot encode QR code: content too large for fixed size QR Code version %d (encoded length is %d bits, maximum length is %d bits)",
			version,
			encoded.Len(),
			chosenVersion.numDataBits())
	}

	q := &QRCode{
		Content:         content,
		Level:           level,
		VersionNumber:   chosenVersion.version,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
		encoder:         encoder,
		data:            encoded,
		version:         *chosenVersion,
	}

	return q, nil
}

// Encode generates a PNG byte slice representing the QR code for the given content, recovery level, and size.
func Encode(content string, level RecoveryLevel, size int) ([]byte, error) {
	q, err := NewQRCode(content, level)
	if err != nil {
		return nil, err
	}

	return q.PNG(size), nil
}

// WriteFile generates a QR code and writes it to a file with the specified filename and size.
// If size is 0, it uses the DefaultFileSize.
func WriteFile(content string, level RecoveryLevel, size int, filename string) error {
	q, err := NewQRCode(content, level)
	if err != nil {
		return err
	}
	if size == 0 {
		return q.WriteFileWithoutSize(filename)
	}
	return q.WriteFile(size, filename)
}

// WriteColorFile generates a QR code with custom background and foreground colors and writes it to a file.
func WriteColorFile(
	content string,
	level RecoveryLevel,
	size int,
	background, foreground color.Color,
	filename string,
) error {
	q, err := NewQRCode(content, level)
	if err != nil {
		return err
	}
	q.BackgroundColor = background
	q.ForegroundColor = foreground

	return q.WriteFile(size, filename)
}

// Bitmap returns a 2D boolean array representing the QR code's modules.
// True represents a foreground module, and false represents a background module.
func (q *QRCode) Bitmap() [][]bool {
	q.encode()

	return q.symbol.bitmap()
}

// Image returns an image.Image representation of the QR code mapped to the specified pixel size.
func (q *QRCode) Image(size int) image.Image {
	q.encode()

	realSize := q.symbol.size

	if size < 0 {
		size = size * -1 * realSize
	}

	if size < realSize {
		size = realSize
	}

	rect := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{size, size}}

	p := color.Palette([]color.Color{q.BackgroundColor, q.ForegroundColor})
	img := image.NewPaletted(rect, p)
	fgClr := uint8(img.Palette.Index(q.ForegroundColor))

	bitmap := q.symbol.bitmap()

	modulesPerPixel := float64(realSize) / float64(size)
	for y := range size {
		y2 := int(float64(y) * modulesPerPixel)
		for x := range size {
			x2 := int(float64(x) * modulesPerPixel)
			v := bitmap[y2][x2]
			if v {
				pos := img.PixOffset(x, y)
				img.Pix[pos] = fgClr
			}
		}
	}

	return img
}

// PNG returns the QR code as a PNG-encoded byte slice mapped to the given size.
func (q *QRCode) PNG(size int) []byte {
	img := q.Image(size)
	buf := new(bytes.Buffer)
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	err := encoder.Encode(buf, img)
	if err != nil {
		return q.PNG(len(buf.Bytes()))
	}

	return buf.Bytes()
}

// Write encodes the QR code as a PNG of the given size and writes it to the provided io.Writer.
func (q *QRCode) Write(size int, out io.Writer) error {
	png := q.PNG(size)
	if _, err := out.Write(png); err != nil {
		return err
	}
	return nil
}

// WriteFile writes the QR code as a PNG to a file with the specified size and filename.
func (q *QRCode) WriteFile(size int, filename string) error {
	return os.WriteFile(filename, q.PNG(size), os.FileMode(0644))
}

// DefaultFileSize defines the standard dimensions used when saving a QR code without a specified size.
const DefaultFileSize = 256

// WriteFileWithoutSize writes the QR code to a file using the DefaultFileSize.
func (q *QRCode) WriteFileWithoutSize(filename string) error {
	return q.WriteFile(DefaultFileSize, filename)
}

// encode processes the QR code data, adds terminators, padding, error correction blocks,
// and determines the optimal masking pattern to build the final symbol.
func (q *QRCode) encode() {
	numTerminatorBits := q.version.numTerminatorBitsRequired(q.data.Len())

	q.addTerminatorBits(numTerminatorBits)
	q.addPadding()

	encoded := q.encodeBlocks()

	const numMasks int = 8
	penalty := 0

	for mask := range numMasks {
		s, err := buildRegularSymbol(q.version, mask, encoded, !q.DisableBorder)
		if err != nil {
			log.Panic(err.Error())
		}

		numEmptyModules := s.numEmptyModules()
		if numEmptyModules != 0 {
			log.Panicf("bug: numEmptyModules is %d (expected 0) (version=%d)",
				numEmptyModules, q.VersionNumber)
		}

		p := s.penaltyScore()

		if q.symbol == nil || p < penalty {
			q.symbol = s
			q.mask = mask
			penalty = p
		}
	}
}

// WithColors sets custom foreground and background colors for the QR code.
// It returns the modified QRCode instance for chaining.
func (q *QRCode) WithColors(foreground, background color.Color) *QRCode {
	q.ForegroundColor = foreground
	q.BackgroundColor = background
	return q
}

// WithNoBorder disables the quiet zone (border) around the QR code.
// It returns the modified QRCode instance for chaining.
func (q *QRCode) WithNoBorder() *QRCode {
	q.DisableBorder = true
	return q
}

// addTerminatorBits appends the necessary terminator bits to the encoded data.
func (q *QRCode) addTerminatorBits(numTerminatorBits int) {
	q.data.AppendNumBools(numTerminatorBits, false)
}

// encodeBlocks splits the data into blocks, calculates error correction codewords for each,
// and interleaves them together into the final bitset.
func (q *QRCode) encodeBlocks() *Bitset {
	type dataBlock struct {
		data          *Bitset
		ecStartOffset int
	}

	block := make([]dataBlock, q.version.numBlocks())

	start := 0
	end := 0
	blockID := 0

	for _, b := range q.version.block {
		for range b.numBlocks {
			start = end
			end = start + b.numDataCodewords*8

			numErrorCodewords := b.numCodewords - b.numDataCodewords
			block[blockID].data = encode(q.data.Substr(start, end), numErrorCodewords)
			block[blockID].ecStartOffset = end - start

			blockID++
		}
	}

	result := New()

	working := true
	for i := 0; working; i += 8 {
		working = false

		for j, b := range block {
			if i >= block[j].ecStartOffset {
				continue
			}

			result.Append(b.data.Substr(i, i+8))

			working = true
		}
	}

	working = true
	for i := 0; working; i += 8 {
		working = false

		for j, b := range block {
			offset := i + block[j].ecStartOffset
			if offset >= block[j].data.Len() {
				continue
			}

			result.Append(b.data.Substr(offset, offset+8))

			working = true
		}
	}

	result.AppendNumBools(q.version.numRemainderBits, false)

	return result
}

// addPadding appends padding codewords to fill the remaining capacity of the chosen version.
func (q *QRCode) addPadding() {
	numDataBits := q.version.numDataBits()

	if q.data.Len() == numDataBits {
		return
	}

	q.data.AppendNumBools(q.version.numBitsToPadToCodeword(q.data.Len()), false)

	padding := [2]*Bitset{
		New(true, true, true, false, true, true, false, false),
		New(false, false, false, true, false, false, false, true),
	}

	i := 0
	for numDataBits-q.data.Len() >= 8 {
		q.data.Append(padding[i])

		i = 1 - i
	}

	if q.data.Len() != numDataBits {
		log.Panicf("BUG: got len %d, expected %d", q.data.Len(), numDataBits)
	}
}

// ToString returns a string representation of the QR code using unicode block characters.
// It is useful for printing the QR code to a standard terminal.
func (q *QRCode) ToString(inverseColor bool) string {
	bits := q.Bitmap()
	var buf bytes.Buffer
	for y := range bits {
		for x := range bits[y] {
			if bits[y][x] != inverseColor {
				buf.WriteString("  ")
			} else {
				buf.WriteString("██")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// ToSmallString returns a more compact string representation of the QR code using half-block characters.
// It is ideal for terminals with limited vertical space.
func (q *QRCode) ToSmallString(inverseColor bool) string {
	bits := q.Bitmap()
	var buf bytes.Buffer
	for y := 0; y < len(bits)-1; y += 2 {
		for x := range bits[y] {
			if bits[y][x] == bits[y+1][x] {
				if bits[y][x] != inverseColor {
					buf.WriteString(" ")
				} else {
					buf.WriteString("█")
				}
			} else {
				if bits[y][x] != inverseColor {
					buf.WriteString("▄")
				} else {
					buf.WriteString("▀")
				}
			}
		}
		buf.WriteString("\n")
	}
	if len(bits)%2 == 1 {
		y := len(bits) - 1
		for x := range bits[y] {
			if bits[y][x] != inverseColor {
				buf.WriteString(" ")
			} else {
				buf.WriteString("▀")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// dataMode defines the type of encoding applied to a specific segment of data.
type dataMode uint8

const (
	dataModeNone dataMode = 1 << iota
	dataModeNumeric
	dataModeAlphanumeric
	dataModeByte
)

// dataEncoderType categorizes QR code versions to apply the correct character count indicators.
type dataEncoderType uint8

const (
	dataEncoderType1To9 dataEncoderType = iota
	dataEncoderType10To26
	dataEncoderType27To40
)

// segment represents a continuous block of data encoded in a single dataMode.
type segment struct {
	dataMode dataMode
	data     []byte
}

// dataEncoder handles the conversion of raw bytes into optimal QR code segments.
type dataEncoder struct {
	minVersion                   int
	maxVersion                   int
	numericModeIndicator         *Bitset
	alphanumericModeIndicator    *Bitset
	byteModeIndicator            *Bitset
	numNumericCharCountBits      int
	numAlphanumericCharCountBits int
	numByteCharCountBits         int
	data                         []byte
	actual                       []segment
	optimised                    []segment
}

// newDataEncoder initializes a dataEncoder configured for a specific version range.
func newDataEncoder(t dataEncoderType) *dataEncoder {
	d := &dataEncoder{}
	black, white := Black, White
	var (
		numericModeIndicator      = New(white, white, white, black)
		alphanumericModeIndicator = New(white, white, black, white)
		byteModeIndicator         = New(white, black, white, white)
	)
	switch t {
	case dataEncoderType1To9:
		d = &dataEncoder{
			minVersion:                   1,
			maxVersion:                   9,
			numericModeIndicator:         numericModeIndicator,
			alphanumericModeIndicator:    alphanumericModeIndicator,
			byteModeIndicator:            byteModeIndicator,
			numNumericCharCountBits:      10,
			numAlphanumericCharCountBits: 9,
			numByteCharCountBits:         8,
		}
	case dataEncoderType10To26:
		d = &dataEncoder{
			minVersion:                   10,
			maxVersion:                   26,
			numericModeIndicator:         numericModeIndicator,
			alphanumericModeIndicator:    alphanumericModeIndicator,
			byteModeIndicator:            byteModeIndicator,
			numNumericCharCountBits:      12,
			numAlphanumericCharCountBits: 11,
			numByteCharCountBits:         16,
		}
	case dataEncoderType27To40:
		d = &dataEncoder{
			minVersion:                   27,
			maxVersion:                   40,
			numericModeIndicator:         numericModeIndicator,
			alphanumericModeIndicator:    alphanumericModeIndicator,
			byteModeIndicator:            byteModeIndicator,
			numNumericCharCountBits:      14,
			numAlphanumericCharCountBits: 13,
			numByteCharCountBits:         16,
		}
	default:
		log.Panic("Unknown dataEncoderType")
	}

	return d
}

// encode processes the raw byte slice into optimized bits suitable for the QR matrix.
func (d *dataEncoder) encode(data []byte) (*Bitset, error) {
	d.data = data
	d.actual = nil
	d.optimised = nil

	if len(data) == 0 {
		return nil, errors.New("no data to encode")
	}

	highestRequiredMode := d.classifyDataModes()

	err := d.optimiseDataModes()
	if err != nil {
		return nil, err
	}

	optimizedLength := 0
	for _, s := range d.optimised {
		length, err := d.encodedLength(s.dataMode, len(s.data))
		if err != nil {
			return nil, err
		}
		optimizedLength += length
	}

	singleByteSegmentLength, err := d.encodedLength(highestRequiredMode, len(d.data))
	if err != nil {
		return nil, err
	}

	if singleByteSegmentLength <= optimizedLength {
		d.optimised = []segment{{dataMode: highestRequiredMode, data: d.data}}
	}

	encoded := New()
	for _, s := range d.optimised {
		d.encodeDataRaw(s.data, s.dataMode, encoded)
	}

	return encoded, nil
}

// classifyDataModes breaks the raw data into actual segments based on the most efficient encoding mode.
func (d *dataEncoder) classifyDataModes() dataMode {
	var start int
	mode := dataModeNone
	highestRequiredMode := mode

	for i, v := range d.data {
		newMode := dataModeNone
		switch {
		case v >= 0x30 && v <= 0x39:
			newMode = dataModeNumeric
		case v == 0x20 || v == 0x24 || v == 0x25 || v == 0x2a || v == 0x2b || v ==
			0x2d || v == 0x2e || v == 0x2f || v == 0x3a || (v >= 0x41 && v <= 0x5a):
			newMode = dataModeAlphanumeric
		default:
			newMode = dataModeByte
		}

		if newMode != mode {
			if i > 0 {
				d.actual = append(d.actual, segment{dataMode: mode, data: d.data[start:i]})

				start = i
			}

			mode = newMode
		}

		if newMode > highestRequiredMode {
			highestRequiredMode = newMode
		}
	}

	d.actual = append(d.actual, segment{dataMode: mode, data: d.data[start:len(d.data)]})

	return highestRequiredMode
}

// optimiseDataModes merges adjacent data segments if coalescing them yields a shorter overall bit sequence.
func (d *dataEncoder) optimiseDataModes() error {
	for i := 0; i < len(d.actual); {
		mode := d.actual[i].dataMode
		numChars := len(d.actual[i].data)

		j := i + 1
		for j < len(d.actual) {
			nextNumChars := len(d.actual[j].data)
			nextMode := d.actual[j].dataMode

			if nextMode > mode {
				break
			}

			coalescedLength, err := d.encodedLength(mode, numChars+nextNumChars)

			if err != nil {
				return err
			}

			seperateLength1, err := d.encodedLength(mode, numChars)

			if err != nil {
				return err
			}

			seperateLength2, err := d.encodedLength(nextMode, nextNumChars)

			if err != nil {
				return err
			}

			if coalescedLength < seperateLength1+seperateLength2 {
				j++
				numChars += nextNumChars
			} else {
				break
			}
		}

		optimised := segment{dataMode: mode,
			data: make([]byte, 0, numChars)}

		for k := i; k < j; k++ {
			optimised.data = append(optimised.data, d.actual[k].data...)
		}

		d.optimised = append(d.optimised, optimised)

		i = j
	}

	return nil
}

// encodeDataRaw takes a segment of bytes and appends its encoded bit representation to the target bitset.
func (d *dataEncoder) encodeDataRaw(data []byte, dataMode dataMode, encoded *Bitset) {
	modeIndicator := d.modeIndicator(dataMode)
	charCountBits := d.charCountBits(dataMode)

	encoded.Append(modeIndicator)

	encoded.AppendUint32(uint32(len(data)), charCountBits)

	switch dataMode {
	case dataModeNumeric:
		for i := 0; i < len(data); i += 3 {
			charsRemaining := len(data) - i

			var value uint32
			bitsUsed := 1

			for j := 0; j < charsRemaining && j < 3; j++ {
				value *= 10
				value += uint32(data[i+j] - 0x30)
				bitsUsed += 3
			}
			encoded.AppendUint32(value, bitsUsed)
		}
	case dataModeAlphanumeric:
		for i := 0; i < len(data); i += 2 {
			charsRemaining := len(data) - i

			var value uint32
			for j := 0; j < charsRemaining && j < 2; j++ {
				value *= 45
				value += encodeAlphanumericCharacter(data[i+j])
			}

			bitsUsed := 6
			if charsRemaining > 1 {
				bitsUsed = 11
			}

			encoded.AppendUint32(value, bitsUsed)
		}
	case dataModeByte:
		for _, b := range data {
			encoded.AppendByte(b, 8)
		}
	}
}

// modeIndicator returns the bits representing the encoding mode for the QR specification.
func (d *dataEncoder) modeIndicator(dataMode dataMode) *Bitset {
	switch dataMode {
	case dataModeNumeric:
		return d.numericModeIndicator
	case dataModeAlphanumeric:
		return d.alphanumericModeIndicator
	case dataModeByte:
		return d.byteModeIndicator
	default:
		log.Panic("Unknown data mode")
	}

	return nil
}

// charCountBits provides the number of bits required to store the length of the data for a given mode.
func (d *dataEncoder) charCountBits(dataMode dataMode) int {
	switch dataMode {
	case dataModeNumeric:
		return d.numNumericCharCountBits
	case dataModeAlphanumeric:
		return d.numAlphanumericCharCountBits
	case dataModeByte:
		return d.numByteCharCountBits
	default:
		log.Panic("Unknown data mode")
	}

	return 0
}

// encodedLength calculates the expected bit length of a segment if it were to be encoded.
func (d *dataEncoder) encodedLength(dataMode dataMode, n int) (int, error) {
	modeIndicator := d.modeIndicator(dataMode)
	charCountBits := d.charCountBits(dataMode)

	if modeIndicator == nil {
		return 0, errors.New("mode not supported")
	}

	maxLength := (1 << uint8(charCountBits)) - 1

	if n > maxLength {
		return 0, errors.New("length too long to be represented")
	}

	length := modeIndicator.Len() + charCountBits

	switch dataMode {
	case dataModeNumeric:
		length += 10 * (n / 3)

		if n%3 != 0 {
			length += 1 + 3*(n%3)
		}
	case dataModeAlphanumeric:
		length += 11 * (n / 2)
		length += 6 * (n % 2)
	case dataModeByte:
		length += 8 * n
	}

	return length, nil
}

// encodeAlphanumericCharacter maps an ASCII byte to its corresponding integer value in the QR alphanumeric table.
func encodeAlphanumericCharacter(v byte) uint32 {
	c := uint32(v)

	switch {
	case c >= '0' && c <= '9':

		return c - '0'
	case c >= 'A' && c <= 'Z':

		return c - 'A' + 10
	case c == ' ':
		return 36
	case c == '$':
		return 37
	case c == '%':
		return 38
	case c == '*':
		return 39
	case c == '+':
		return 40
	case c == '-':
		return 41
	case c == '.':
		return 42
	case c == '/':
		return 43
	case c == ':':
		return 44
	default:
		log.Panicf("encodeAlphanumericCharacter() with non alphanumeric char %v.", v)
	}

	return 0
}
