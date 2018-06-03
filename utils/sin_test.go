package utils_test

import (
	"testing"
	"math"
	"github.com/20zinnm/spac/utils"
	"math/rand"
	"time"
	"fmt"
)

var (
	SinTrials = 1000000
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestFastSin(t *testing.T) {
	thetas := make([]float64, 0, SinTrials)
	for i := 0; i < SinTrials; i++ {
		thetas = append(thetas, rand.ExpFloat64()/.1)
	}
	errors := make([]float64, 0, SinTrials)
	for _, theta := range thetas {
		fast := utils.FastSin(theta)
		actual := math.Sin(theta)
		errors = append(errors, math.Abs((fast-actual)/actual))
	}
	var average float64
	for _, err := range errors {
		average += err / float64(SinTrials)
	}
	fmt.Printf("average error: %f%%\n", average*.01)
}

func BenchmarkFastSin(b *testing.B) {
	theta := 30 * math.Pi / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.FastSin(theta)
	}
}

func BenchmarkStdSin(b *testing.B) {
	theta := 30 * math.Pi / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		math.Sin(theta)
	}
}
