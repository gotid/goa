package util

import (
	"bytes"
	goformat "go/format"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/z-sdk/goa/lib/fs"
)

type TemplateX struct {
	name     string
	text     string
	goFmt    bool
	savePath string
}

func With(name string) *TemplateX {
	return &TemplateX{
		name: name,
	}
}

func (t *TemplateX) Parse(text string) *TemplateX {
	t.text = text
	return t
}

func (t *TemplateX) GoFmt(format bool) *TemplateX {
	t.goFmt = format
	return t
}

func (t *TemplateX) SaveTo(data interface{}, path string, forceUpdate bool) error {
	if fs.FileExist(path) && !forceUpdate {
		return nil
	}
	output, err := t.execute(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, output.Bytes(), os.ModePerm)
}

func (t *TemplateX) Execute(data interface{}) (*bytes.Buffer, error) {
	return t.execute(data)
}

func (t *TemplateX) execute(data interface{}) (*bytes.Buffer, error) {
	tpl, err := template.New(t.name).Parse(t.text)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	if !t.goFmt {
		return buf, nil
	}
	output, err := goformat.Source(buf.Bytes())
	if err != nil {
		return nil, err
	}
	buf.Reset()
	buf.Write(output)
	return buf, nil
}
