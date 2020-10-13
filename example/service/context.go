package service

import (
	model "github.com/z-sdk/goa/example/model/label"
	"github.com/z-sdk/goa/lib/store/cache"
	"github.com/z-sdk/goa/lib/store/redis"
	"github.com/z-sdk/goa/lib/store/sqlx"
)

type Config struct {
	DataSource string
	Table      string
	Cache      cache.ClusterConf
}

type Context struct {
	c        Config
	TagModel *model.TagModel
}

func NewServiceContext() *Context {
	c := Config{
		DataSource: "root:asdfasdf@tcp(192.168.0.166:3306)/nest_user?parseTime=true",
		Table:      "nest_label.tag",
		Cache: cache.ClusterConf{
			{
				Conf: redis.Conf{
					Host: "192.168.0.166:6800",
					Mode: redis.StandaloneMode,
				},
				Weight: 100,
			},
		},
	}

	return &Context{
		c:        c,
		TagModel: model.NewTagModel(sqlx.NewMySQL(c.DataSource), c.Cache, c.Table),
	}
}
