package convjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Value describes a value
type Value struct {
	typ ValueType
	val reflect.Value
}

// ValueType describes value type
type ValueType int8

// all available value type
const (
	TypeNil         = ValueType(0)
	TypeInt         = ValueType(1)
	TypeUint        = ValueType(2)
	TypeFloat       = ValueType(3)
	TypeBool        = ValueType(4)
	TypeString      = ValueType(5)
	TypeArray       = ValueType(6)
	TypeMap         = ValueType(7)
	TypeStruct      = ValueType(8)
	TypeUnsupported = ValueType(-1)
)

type (
	// Map map type value
	Map = map[string]interface{}
	// Array array type value
	Array = []interface{}
)

var (
	// Nil represents a nil value
	Nil = NewValue(nil)
)

// NewValue new value
func NewValue(val interface{}) *Value {
	if val == nil {
		return newValue(TypeNil, reflect.Value{})
	}

	// get the value that val points to
	v := reflect.Indirect(reflect.ValueOf(val))
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return newValue(TypeInt, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return newValue(TypeUint, v)
	case reflect.Float32, reflect.Float64:
		return newValue(TypeFloat, v)
	case reflect.Bool:
		return newValue(TypeBool, v)
	case reflect.String:
		return newValue(TypeString, v)
	case reflect.Array, reflect.Slice:
		return newValue(TypeArray, v)
	case reflect.Map:
		return newValue(TypeMap, v)
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(Value{}) {
			return val.(*Value)
		}
		return newValue(TypeMap, reflect.ValueOf(convert2Map(val)))
	default:
		return newValue(TypeUnsupported, reflect.Value{})
	}
}

func newValue(valueType ValueType, value reflect.Value) *Value {
	return &Value{
		typ: valueType,
		val: value,
	}
}

func convert2Map(val interface{}) map[string]interface{} {
	data, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal(data, &res); err != nil {
		panic(err)
	}

	return res
}

// String returns value as string
func (v Value) String() (string, error) {
	switch v.typ {
	case TypeNil:
		return "", nil
	case TypeInt:
		return strconv.FormatInt(v.val.Int(), 10), nil
	case TypeUint:
		return strconv.FormatUint(v.val.Uint(), 10), nil
	case TypeFloat:
		return strconv.FormatFloat(v.val.Float(), 'e', -1, 64), nil
	case TypeBool:
		return strconv.FormatBool(v.val.Bool()), nil
	case TypeString:
		return v.val.String(), nil
	case TypeMap, TypeStruct, TypeArray:
		return v.JSON()
	default:
		return "", errors.New(fmt.Sprintf("Failed to convert value with type %s into TypeString", v.val.Type()))
	}
}

// MustString returns string value, an empty string will return if
// anything wrong occurs.
func (v Value) MustString() string {
	val, _ := v.String()
	return val
}

// Int returns value as int64 type
func (v Value) Int() (int64, error) {
	switch v.typ {
	case TypeNil:
		return 0, nil
	case TypeInt:
		return v.val.Int(), nil
	case TypeUint:
		return int64(v.val.Uint()), nil
	case TypeFloat:
		return int64(v.val.Float()), nil
	case TypeBool:
		if v.val.Bool() {
			return 1, nil
		}
		return 0, nil
	case TypeString:
		return strconv.ParseInt(v.val.String(), 10, 64)
	default:
		return 0, errors.New(fmt.Sprintf("Failed to convert value with type %s into TypeInt", v.val.Type()))
	}
}

// MustInt returns value as int type ignoring error
func (v Value) MustInt() int64 {
	val, _ := v.Int()
	return val
}

// Uint returns value as unsigned int type
func (v Value) Uint() (uint64, error) {
	value, err := v.Int()
	return uint64(value), err
}

// MustUint returns value as unsigned int type ignoring error
func (v Value) MustUint() uint64 {
	value, _ := v.Int()
	return uint64(value)
}

// Float returns value as float64 type
func (v Value) Float() (float64, error) {
	switch v.typ {
	case TypeNil:
		return 0, nil
	case TypeInt:
		return float64(v.val.Int()), nil
	case TypeUint:
		return float64(v.val.Uint()), nil
	case TypeFloat:
		return v.val.Float(), nil
	case TypeBool:
		if v.val.Bool() {
			return float64(1), nil
		}
		return float64(0), nil
	case TypeString:
		return strconv.ParseFloat(v.val.String(), 64)
	default:
		return float64(0), errors.New(fmt.Sprintf("Failed to convert value with type %s into TypeFloat", v.val.Type()))
	}
}

// MustFloat returns value as float64 type ignoring error
func (v Value) MustFloat() float64 {
	val, _ := v.Float()
	return float64(val)
}

// Bool returns value as boolean type
func (v Value) Bool() (bool, error) {
	switch v.typ {
	case TypeNil:
		return false, nil
	case TypeInt:
		return v.val.Int() == 1, nil
	case TypeUint:
		return v.val.Uint() == 1, nil
	case TypeFloat:
		return v.val.Float() == 1, nil
	case TypeBool:
		return v.val.Bool(), nil
	case TypeString:
		return v.val.String() == "true", nil
	default:
		return false, errors.New(fmt.Sprintf("failed to convert value with type %s into BoolType", v.val.Type()))
	}
}

// MustBool returns value as boolean type ignoring error
func (v Value) MustBool() bool {
	val, _ := v.Bool()
	return val
}

// JSON json string
func (v Value) JSON() (string, error) {
	if v.typ == TypeStruct || v.typ == TypeMap || v.typ == TypeArray {
		b, err := json.Marshal(v.val.Interface())
		return string(b), err
	}
	return "", errors.New("Failed to serialize into json")
}

// IsNil returns true if value represents nil
func (v Value) IsNil() bool {
	return v.typ == TypeNil
}

// Set binds value to path.
func (v *Value) Set(path string, value interface{}, options ...OptionFunc) error {
	if path == "" {
		temp := NewValue(value)
		v.typ = temp.typ
		v.val = temp.val
		return nil
	}

	// TODO: check whether it's a valid type
	opt := newOption()
	opt.load(options...)

	keys := strings.Split(path, opt.delimiter)
	key := keys[0]
	if v.typ != TypeMap {
		return errors.Errorf("path %s is not available", path)
	}

	mdata := v.val.Interface().(Map)

	// TODO: support set array element
	keyname, _, err := parseKey(key)
	if err != nil {
		return err
	}

	val := NewValue(mdata[keyname])
	val.Set(strings.Join(keys[1:], "."), value)

	mdata[keyname] = val

	v.val = reflect.ValueOf(mdata)

	return nil
}

// Get get value by path.
func (v Value) Get(path string, options ...OptionFunc) (*Value, error) {
	opt := newOption()
	opt.load(options...)

	keys := strings.Split(path, opt.delimiter)
	cur := &v
	for _, key := range keys {
		if cur.typ != TypeMap {
			return Nil, errors.Errorf("path %s is not available", path)
		}

		mdata := cur.val.Interface().(Map)

		keyname, indexes, err := parseKey(key)
		if err != nil {
			return Nil, err
		}

		val := mdata[keyname]
		cur = NewValue(val)

		var curPath string = keyname
		for _, index := range indexes {
			array, ok := cur.val.Interface().(Array)
			if !ok {
				return Nil, fmt.Errorf("invalid path: %s is not an array", curPath)
			}

			elem := array[index]
			cur = NewValue(elem)

			curPath = fmt.Sprintf("%s[%v]", curPath, index)
		}
	}

	return cur, nil
}

var keyPattern = regexp.MustCompile(`^[a-zA-Z_@][a-zA-Z0-9-_.]{0,}(\[[0-9]+\]){0,}$`)

// parseKey parse given key to obtain true keyname and indexes if the key represents
// an array. For instance, parsing key 'a[1][2]' will return keyname 'a' and indexes '[]int{1, 2}'
func parseKey(key string) (keyname string, indexes []int, err error) {
	if !keyPattern.Match([]byte(key)) {
		err = fmt.Errorf("invalid key: doesn't match '%s'", keyPattern.String())
		return
	}

	indx := strings.Index(key, "[")
	if indx < 0 {
		keyname = key
		return
	}

	keyname = key[:indx]
	var numStr string
	for i := indx; i < len(key); i++ {
		if key[i] == '[' {
			continue
		}

		if key[i] == ']' {
			index, _ := strconv.ParseInt(numStr, 10, 64)
			indexes = append(indexes, int(index))
			numStr = ""
			continue
		}
		numStr = numStr + string(key[i])
	}
	return
}

// MarshalJSON implements Marshaler interface.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val.Interface())
}
