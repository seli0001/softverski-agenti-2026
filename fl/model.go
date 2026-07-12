package fl

import "math"

func Train(X [][]float64, y []float64, w []float64, epochs int, lr float64) []float64 {
	for e := 0; e < epochs; e++ {
		for i := range X {
			z := w[len(w)-1]
			for j := range X[i] {
				z += w[j] * X[i][j]
			}
			p := 1.0 / (1.0 + math.Exp(-z))
			err := p - y[i]
			for j := range X[i] {
				w[j] -= lr * err * X[i][j]
			}
			w[len(w)-1] -= lr * err

		}
	}
	return w
}

func Accuracy(X [][]float64, y []float64, w []float64) float64 {
	hits := 0
	for i := range X {
		z := w[len(w)-1]
		for j := range X[i] {
			z += w[j] * X[i][j]
		}
		p := 1.0 / (1.0 + math.Exp(-z))
		if (p > 0.5) == (y[i] == 1.0) {
			hits++
		}
	}
	return float64(hits) / float64(len(y))
}
