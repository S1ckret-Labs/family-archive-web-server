package helpers

func ZipToMap[K comparable, V any](a []K, b []V) map[K]V {
	c := make(map[K]V)
	for i := 0; i < len(a); i++ {
		c[a[i]] = b[i]
	}
	return c
}
