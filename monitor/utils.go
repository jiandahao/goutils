package monitor

import "reflect"

func assertNotNil(value interface{}, desc string) {
	v := reflect.ValueOf(value)
	if v.IsNil() {
		panic(desc)
	}
}
