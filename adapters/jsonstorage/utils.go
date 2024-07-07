package jsonstorage

func sliceWithout(slice []string, value string) []string {
	t := 0
	for _, value2 := range slice {
		if value == value2 {
			continue
		}

		slice[t] = value2
		t += 1
	}

	return slice[:t]
}
