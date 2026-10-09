package vector

import (
	"errors"
	"math"
)

var ErrDimensionsMismatch = errors.New("vector dimensions mismatch")
var ErrZeroMagnitude = errors.New("vector has zero magnitude")

func DotProduct(a, b []float32) (q float32, err error) {
	if len(a) != len(b) {
		return 0, ErrDimensionsMismatch
	}
	var sum float32
	for i := 0 ; i < len(a) ; i++ {
		sum += float32(a[i] * b[i])
	}
	return sum, nil
}

func Magnitude(a []float32) float32 {
	var sum float32;
	for i := 0 ; i < len(a) ; i++ {
		sum += a[i] * a[i]
	}
	return float32(math.Sqrt(float64(sum)))
}

func CosineSimiliarity(a, b []float32) (float32, error) {
	dot , err := DotProduct(a, b)
	if err != nil {
		return 0, err
	}
	magA := Magnitude(a)
	magB := Magnitude(b)
	if magA == 0 || magB == 0 {
		return 0, ErrZeroMagnitude
	}
	return dot / (magA * magB), nil	
}