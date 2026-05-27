package utils

func MapToGql[T comparable](colors []T) ([]*T, error) {
	colorsChan := make(chan *T, len(colors))
	var colorsPtrs []*T
	go func() {
		for i := range colors {
			colorsChan <- &colors[i]
		}
		close(colorsChan)
	}()
	for category := range colorsChan {
		colorsPtrs = append(colorsPtrs, category)
	}
	return colorsPtrs, nil
}
