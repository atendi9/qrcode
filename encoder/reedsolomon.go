package encoder

import (
	"fmt"
	"log"
)

// encode applies Reed-Solomon error correction to the provided data.
// It calculates the error correction codewords based on the specified numECBytes
// and appends them to a clone of the original data Bitset, returning the result.
func encode(data *Bitset, numECBytes int) *Bitset {
	ecpoly := newGFPolyFromData(data)
	ecpoly = polyOperator{}.Multiply(ecpoly, newGFPolyMonomial(gfOne, numECBytes))
	generator := generatorPoly(numECBytes)
	remainder := polyOperator{}.Remainder(ecpoly, generator)
	result := Clone(data)
	result.AppendBytes(remainder.data(numECBytes))

	return result
}

// polyOperator provides mathematical operations for Galois Field polynomials (gfPoly).
type polyOperator struct{}

// NewGFPolyOperator creates and returns a new polyOperator instance.
func NewGFPolyOperator() polyOperator {
	return polyOperator{}
}

// Multiply multiplies two Galois Field polynomials (a and b) and returns the resulting gfPoly.
func (op polyOperator) Multiply(a, b gfPoly) gfPoly {
	numATerms := a.numTerms()
	numBTerms := b.numTerms()

	result := gfPoly{term: make([]gfElement, numATerms+numBTerms)}

	for i := range numATerms {
		for j := range numBTerms {
			if a.term[i] != 0 && b.term[j] != 0 {
				monomial := gfPoly{term: make([]gfElement, i+j+1)}
				monomial.term[i+j] = elementOperator{a.term[i], b.term[j]}.Multiply()

				result = op.Add(result, monomial)
			}
		}
	}

	return result.normalised()
}

// Remainder calculates the remainder of the division between a numerator and a denominator gfPoly.
// It panics if the denominator is a zero polynomial.
func (op polyOperator) Remainder(numerator, denominator gfPoly) gfPoly {
	if denominator.equals(gfPoly{}) {
		log.Panicln("Remainder by zero")
	}

	remainder := numerator

	for remainder.numTerms() >= denominator.numTerms() {
		degree := remainder.numTerms() - denominator.numTerms()
		coefficient := elementOperator{
			remainder.term[remainder.numTerms()-1],
			denominator.term[denominator.numTerms()-1],
		}.Divide()

		divisor := op.Multiply(denominator,
			newGFPolyMonomial(coefficient, degree))

		remainder = op.Add(remainder, divisor)
	}

	return remainder.normalised()
}

// Add adds two Galois Field polynomials (a and b) together and returns the resulting gfPoly.
func (op polyOperator) Add(a, b gfPoly) gfPoly {
	numATerms := a.numTerms()
	numBTerms := b.numTerms()

	numTerms := numATerms
	if numBTerms > numTerms {
		numTerms = numBTerms
	}

	result := gfPoly{term: make([]gfElement, numTerms)}

	for i := range numTerms {
		switch {
		case numATerms > i && numBTerms > i:
			result.term[i] = elementOperator{a.term[i], b.term[i]}.Add()
		case numATerms > i:
			result.term[i] = a.term[i]
		default:
			result.term[i] = b.term[i]
		}
	}

	return result.normalised()
}

// newGFPolyFromData constructs a gfPoly from a given Bitset.
// It groups the bits into bytes to form the polynomial coefficients.
func newGFPolyFromData(data *Bitset) gfPoly {
	numTotalBytes := data.Len() / 8
	if data.Len()%8 != 0 {
		numTotalBytes++
	}

	result := gfPoly{term: make([]gfElement, numTotalBytes)}

	i := numTotalBytes - 1
	for j := 0; j < data.Len(); j += 8 {
		result.term[i] = gfElement(data.ByteAt(j))
		i--
	}

	return result
}

// newGFPolyMonomial creates a monomial gfPoly with the given term (coefficient) and degree.
// It returns an empty polynomial if the term is gfZero.
func newGFPolyMonomial(term gfElement, degree int) gfPoly {
	if term == gfZero {
		return gfPoly{}
	}

	result := gfPoly{term: make([]gfElement, degree+1)}
	result.term[degree] = term

	return result
}

// gfPoly represents a polynomial over a Galois Field.
type gfPoly struct {
	term []gfElement
}

// generatorPoly generates a Reed-Solomon generator polynomial for the specified degree.
// It panics if the requested degree is less than 2.
func generatorPoly(degree int) gfPoly {
	if degree < 2 {
		log.Panic("degree < 2")
	}

	gen := gfPoly{term: []gfElement{1}}

	for i := range degree {
		nextPoly := gfPoly{term: []gfElement{gfExpTable[i], 1}}
		gen = polyOperator{}.Multiply(gen, nextPoly)
	}

	return gen
}

// numTerms returns the number of terms (coefficients) in the polynomial.
func (e gfPoly) numTerms() int {
	return len(e.term)
}

// normalised removes any trailing zero coefficients from the polynomial
// to ensure its internal length matches its actual mathematical degree.
func (e gfPoly) normalised() gfPoly {
	numTerms := e.numTerms()
	maxNonzeroTerm := numTerms - 1

	for i := numTerms - 1; i >= 0; i-- {
		if e.term[i] != 0 {
			break
		}

		maxNonzeroTerm = i - 1
	}

	if maxNonzeroTerm < 0 {
		return gfPoly{}
	} else if maxNonzeroTerm < numTerms-1 {
		e.term = e.term[0 : maxNonzeroTerm+1]
	}

	return e
}

// data converts the polynomial coefficients into a byte slice of length numTerms.
func (e gfPoly) data(numTerms int) []byte {
	result := make([]byte, numTerms)

	i := numTerms - len(e.term)
	for j := len(e.term) - 1; j >= 0; j-- {
		result[i] = byte(e.term[j])
		i++
	}

	return result
}

// string returns a string representation of the polynomial.
// If useIndexForm is true, it formats coefficients as powers of the generator element (alpha).
func (e gfPoly) string(useIndexForm bool) string {
	var str string
	numTerms := e.numTerms()

	for i := numTerms - 1; i >= 0; i-- {
		if e.term[i] > 0 {
			if len(str) > 0 {
				str += " + "
			}

			if !useIndexForm {
				str += fmt.Sprintf("%dx^%d", e.term[i], i)
			} else {
				str += fmt.Sprintf("a^%dx^%d", gfLogTable[e.term[i]], i)
			}
		}
	}

	if len(str) == 0 {
		str = "0"
	}

	return str
}

// equals evaluates whether two polynomials are mathematically identical.
func (e gfPoly) equals(other gfPoly) bool {
	var minecPoly *gfPoly
	var maxecPoly *gfPoly

	if e.numTerms() > other.numTerms() {
		minecPoly = &other
		maxecPoly = &e
	} else {
		minecPoly = &e
		maxecPoly = &other
	}

	numMinTerms := minecPoly.numTerms()
	numMaxTerms := maxecPoly.numTerms()

	for i := range numMinTerms {
		if e.term[i] != other.term[i] {
			return false
		}
	}

	for i := numMinTerms; i < numMaxTerms; i++ {
		if maxecPoly.term[i] != 0 {
			return false
		}
	}

	return true
}

// gfElement represents an individual element in the Galois Field GF(2^8).
type gfElement uint8

const (
	// gfZero represents the zero element in the Galois Field.
	gfZero = gfElement(0)
	// gfOne represents the identity element in the Galois Field.
	gfOne = gfElement(1)
)

// Element converts a standard byte value into a Galois Field element.
func Element(data byte) gfElement {
	return gfElement(data)
}

// elementOperator provides mathematical operations for individual Galois Field elements.
type elementOperator struct {
	a, b gfElement
}

// Add performs Galois Field addition (bitwise XOR) on the elements a and b.
func (op elementOperator) Add() gfElement {
	return op.a ^ op.b
}

// Sub performs Galois Field subtraction (bitwise XOR) on the elements a and b.
// Note that in GF(2^8), addition and subtraction are equivalent operations.
func (op elementOperator) Sub() gfElement {
	return op.a ^ op.b
}

// Multiply computes the product of elements a and b using exponential and logarithm tables.
func (op elementOperator) Multiply() gfElement {
	if op.a == gfZero || op.b == gfZero {
		return gfZero
	}

	return gfExpTable[(gfLogTable[op.a]+gfLogTable[op.b])%255]
}

// Divide divides element a by element b. It panics if the divisor (b) is zero.
func (op elementOperator) Divide() gfElement {
	if op.a == gfZero {
		return gfZero
	} else if op.b == gfZero {
		log.Panicln("Divide by zero")
	}

	inv := op.Inverse(op.b)
	return elementOperator{a: op.a, b: inv}.Multiply()
}

// Inverse calculates the multiplicative inverse of element a.
// It panics if a is zero, as zero has no multiplicative inverse.
func (op elementOperator) Inverse(a gfElement) gfElement {
	if a == gfZero {
		log.Panicln("No multiplicative inverse of 0")
	}

	return gfExpTable[255-gfLogTable[a]]
}
