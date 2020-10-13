package service

import (
	model "github.com/z-sdk/goa/example/model/label"
)

func FindOne(id int64) (*model.Tag, error) {
	ctx := NewServiceContext()
	return ctx.TagModel.FindOne(id)
}

func GetNames(ids []int64) []string {
	names := make([]string, len(ids))
	for i := 0; i < len(ids); i++ {
		one, err := FindOne(ids[i])
		if err == nil {
			names = append(names, one.Name)
		} else {
			names = append(names, "")
		}
	}
	return names
}
