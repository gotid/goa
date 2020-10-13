package cache

import "strings"

const keySeparator = ","

func TotalWeights(configs []Conf) int {
	var weights int

	for _, conf := range configs {
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
