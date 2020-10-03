package mapping

import (
	"fmt"
	"reflect"
	"strconv"
)

// Repr 返回对应的字符串表达形式
func Repr(v interface{}) string {
	if v == nil {
		return ""
	}

	// 字符串类型判断
	switch vt := v.(type) {
	case fmt.Stringer:
		return vt.String()
	}

	//if _, ok := v.(string); ok {
	//	return fmt.Sprint(v)
	//}

	//if reflect.TypeOf(v).Name() == "string" {
	//	return fmt.Sprint(v)
	//}

	// 指针类型处理
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	// 根据值接口的类型进行处理
	switch vt := val.Interface().(type) {
	case bool:
		return strconv.FormatBool(vt)
	case error:
		return vt.Error()
	case float32:
		return strconv.FormatFloat(float64(vt), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(vt, 'f', -1, 32)
	case fmt.Stringer:
		return vt.String()
	case int:
		return strconv.Itoa(vt)
	case int8:
		return strconv.Itoa(int(vt))
	case int16:
		return strconv.Itoa(int(vt))
	case int32:
		return strconv.Itoa(int(vt))
	case int64:
		return strconv.FormatInt(vt, 10)
	case string:
		return vt
	case uint:
		return strconv.FormatUint(uint64(vt), 10)
	case uint8:
		return strconv.FormatUint(uint64(vt), 10)
	case uint16:
		return strconv.FormatUint(uint64(vt), 10)
	case uint32:
		return strconv.FormatUint(uint64(vt), 10)
	case uint64:
		return strconv.FormatUint(vt, 10)
	case []byte:
		return string(vt)
	default:
		return fmt.Sprint(val.Interface())
	}
}

// ValidatePtr 验证值是否为有效的指针，无效则报错
func ValidatePtr(v *reflect.Value) error {
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("非有效指针: %v", v)
	}
	return nil
}

// Deref 解引用：取值（如是指针，则返回值）
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
