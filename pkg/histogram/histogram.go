package histogram

type Histogram struct {
	Values [256]int
}

func (a Histogram) At(x int) int {
	return (a.Values[x])
}

func (hist Histogram) XY(x int) (a, b float64) {
	return float64(x), float64(hist.At(x))
}

func (hist Histogram) Len() int {
	return 255
}

type HistogramNormalized struct {
	Values [256]float64
}

func (a HistogramNormalized) At(x int) float64 {
	return (a.Values[x])
}
