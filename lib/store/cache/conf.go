package cache

import "goa/lib/store/redis"

type (
	Conf struct {
		redis.Conf
		Weight int `json:",default=100"`
	}

	ClusterConf []Conf
)
