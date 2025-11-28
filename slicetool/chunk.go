package slicetool

func Chunk[T any](slice []T, batchSize int) [][]T {
	var batches [][]T

	for i := 0; i < len(slice); i += batchSize {
		end := i + batchSize
		if end > len(slice) {
			end = len(slice)
		}
		batches = append(batches, slice[i:end])
	}

	return batches
}
