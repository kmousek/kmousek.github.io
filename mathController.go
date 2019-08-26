package controller

import "math"

func RoundOff (val float64, idx int) float64 {
	var f64Result float64
	var f64Half float64
	var f64Factor float64

	f64Half = 0.5
	f64Factor = 1

	for i:=0;i<idx;i++{
		f64Half *= 0.1
		f64Factor *= 10
	}

	f64Result = math.Trunc(((val+f64Half)*f64Factor))/f64Factor
	return f64Result
}