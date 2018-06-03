package utils

import "math"

const (
	InversePi = 1 / math.Pi
)

func FastSin(x float64) float64 {
	k := int(x * InversePi)
	x -= float64(k) * math.Pi
	x2 := x * x
	x = x * (0.99969198629596757779830113868360584 + x2*(-0.16528911397014738207016302002888890+0.00735246819687011731341356165096815*x2));
	if k%2 == 1 {
		x = -x
	}
	return x
}
