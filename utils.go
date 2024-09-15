package main

func clamp(v, min_v, max_v int) int {
	if v < min_v {
		return min_v
	}
	if v > max_v {
		return max_v
	}
	return v
}
