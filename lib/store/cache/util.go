package cache

import "strings"

const keySeparator = ","

func TotalWeights(confs []Conf) int {
	var weights int

	for _, conf := range confs {
		if conf.Weight < 0 {
			conf.Weight = 0
		}
		weights += conf.Weight
	}

	return weights
}

func formatKeys(keys []string) string {
	return strings.Join(keys, keySeparator)
}
