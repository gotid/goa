package redis

import (
	"github.com/stretchr/testify/assert"
	"github.com/z-sdk/goa/lib/stringx"
	"testing"
)

func TestConf_Validate(t *testing.T) {
	tests := []struct {
		name string
		Conf
		ok bool
	}{
		{
			name: "缺失主机",
			Conf: Conf{
				Host:     "",
				Mode:     StandaloneMode,
				Password: "",
			},
			ok: false,
		},
		{
			name: "缺失类型",
			Conf: Conf{
				Host:     "localhost:6379",
				Mode:     "",
				Password: "",
			},
			ok: false,
		},
		{
			name: "正常",
			Conf: Conf{
				Host:     "localhost:6379",
				Mode:     StandaloneMode,
				Password: "",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(stringx.RandId(), func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.Conf.Validate())
			} else {
				assert.NotNil(t, test.Conf.Validate())
			}
		})
	}
}
func TestKeyConf_Validate(t *testing.T) {
	tests := []struct {
		name string
		KeyConf
		ok bool
	}{
		{
			name: "缺失主机",
			KeyConf: KeyConf{
				Conf: Conf{
					Host:     "",
					Mode:     StandaloneMode,
					Password: "",
				},
				Key: "",
			},
			ok: false,
		},
		{
			name: "缺失类型",
			KeyConf: KeyConf{
				Conf: Conf{
					Host:     "localhost:6379",
					Mode:     "",
					Password: "",
				},
				Key: "",
			},
			ok: false,
		},
		{
			name: "正常",
			KeyConf: KeyConf{
				Conf: Conf{
					Host:     "localhost:6379",
					Mode:     StandaloneMode,
					Password: "",
				},
				Key: "foo",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(stringx.RandId(), func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.KeyConf.Validate())
			} else {
				assert.NotNil(t, test.KeyConf.Validate())
			}
		})
	}
}
