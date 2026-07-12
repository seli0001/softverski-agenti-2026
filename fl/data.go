package fl

import (
	"encoding/csv"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

func Load(path string) ([][]float64, []float64) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)

	var X [][]float64
	var y []float64
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		t, err1 := strconv.ParseFloat(strings.TrimSpace(rec[5]), 64)
		if err1 != nil {
			continue
		}
		dur, err2 := strconv.ParseFloat(strings.TrimSpace(rec[6]), 64)
		if err2 != nil {
			continue
		}
		delay, err3 := strconv.ParseFloat(strings.TrimSpace(rec[7]), 64)
		if err3 != nil {
			continue
		}
		day, err := strconv.Atoi(strings.TrimSpace(rec[4]))
		if err != nil {
			continue
		}

		t = t / 1440.0
		dur = dur / 700.0
		row := []float64{t, dur}
		days := make([]float64, 7)
		days[day-1] = 1.0
		row = append(row, days...)
		X = append(X, row)
		y = append(y, delay)
	}
	rand.Shuffle(len(X), func(i, j int) {
		X[i], X[j] = X[j], X[i]
		y[i], y[j] = y[j], y[i]
	})
	return X, y
}

func Slice(X [][]float64, y []float64, index, size int) ([][]float64, []float64) {
	start := index * size
	end := start + size
	if end > len(X) {
		end = len(X)
	}
	return X[start:end], y[start:end]
}

func TestSet(X [][]float64, y []float64, size int) ([][]float64, []float64) {
	return X[len(X)-size:], y[len(y)-size:]
}
