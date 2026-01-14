package wmimage

import (
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

func blockEnergy(block [8][8]float64) float64 {
	data := make([]float64, 64)
	i := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			data[i] = block[y][x]
			i++
		}
	}
	v := mat.NewVecDense(64, data)
	return mat.Norm(v, 2)
}

func randomBlockOrder(n int, seed int64) []int {
	r := rand.New(rand.NewSource(seed))
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		idx[i] = i
	}
	r.Shuffle(n, func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	return idx
}
