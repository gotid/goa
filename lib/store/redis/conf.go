package redis

import "errors"

var (
	ErrEmptyHost = errors.New("redis host 为空")
	ErrEmptyMode = errors.New("redis mode 为空")
	ErrEmptyKey  = errors.New("redis key 为空")
)

type (
	Conf struct {
		Host     string
		Mode     string `json:",default=standalone,options=standalone|cluster"` // 默认单点模式，可选集群模式
		Password string `json:",optional"`
	}

	KeyConf struct {
		Conf
		Key string `json:",optional"`
	}
)

func (c Conf) NewRedis() *Redis {
	return NewRedis(c.Host, c.Mode, c.Password)
}

func (c Conf) Validate() error {
	if len(c.Host) == 0 {
		return ErrEmptyHost
	}

	if len(c.Mode) == 0 {
		return ErrEmptyMode
	}

	return nil
}

func (kc KeyConf) Validate() error {
	if err := kc.Conf.Validate(); err != nil {
		return err
	}

	if len(kc.Key) == 0 {
		return ErrEmptyKey
	}

	return nil
}
