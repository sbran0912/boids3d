package lib3d

import (
	"math"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func random(start, end int) int {
	return rng.Intn(end-start+1) + start
}

func RandomFloat(start, end float32) float32 {
	scale := rng.Float32()
	return start + scale*(end-start)
}

func limitNum(number, limit float32) float32 {
	vorzeichen := float32(-1)
	if number >= 0 {
		vorzeichen = 1
	}
	numberMag := number * vorzeichen
	if numberMag > limit {
		numberMag = limit
	}
	return numberMag * vorzeichen
}

func ConstrainNum(value, min, max float32) float32 {
	return float32(math.Min(float64(max), math.Max(float64(min), float64(value))))
}

func mapNum(n, rangeOld, rangeNew float32) float32 {
	return n / rangeOld * rangeNew
}
