package errorx

import "bytes"

type (
	Errors struct {
		errs errArray
	}
	errArray []error
)

func (es errArray) Error() string {
	var buf bytes.Buffer

	for i := range es {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(es[i].Error())
	}

	return buf.String()
}

// Add 追加一条错误
func (es *Errors) Add(err error) {
	if err != nil {
		es.errs = append(es.errs, err)
	}
}

func (es *Errors) NotNil() bool {
	return len(es.errs) > 0
}

func (es *Errors) Error() error {
	switch len(es.errs) {
	case 0:
		return nil
	case 1:
		return es.errs[0]
	default:
		return es.errs
	}
}
