package histogram

type Histogram struct {
	values [256]int
}

func (a Histogram) At(x int) int {
	if x >= 0 && x <= 255 {
		return (a.values[x])
	}
	return -1
}

func (hist Histogram) XY(x int) (a, b float64) {
	return float64(x), float64(hist.At(x))
}

func (hist Histogram) Len() int {
	return 255
}
