package core

import "math"

func TrimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

//type MapTickers struct {
//	tickers map[string]Ticker
//}
//
//func (b MapTickers) copy() MapTickers {
//	tickers := map[string]Ticker{}
//	for k, v := range b.tickers {
//		tickers[k] = v
//	}
//	return  MapTickers{tickers}
//}

func Round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func Trunc(x, unit float64) float64 {
	return math.Trunc(x/unit) * unit
}


func RoundDown(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	newVal = round / pow
	return
}