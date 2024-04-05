package slicex

func DistinctUint64Slice(input []uint64) []uint64 {
	m := make(map[uint64]struct{})
	for _, v := range input {
		m[v] = struct{}{}
	}

	output := make([]uint64, 0, len(m))
	for i := range m {
		output = append(output, i)
	}
	return output
}
