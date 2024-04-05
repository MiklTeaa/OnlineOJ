package strconvx

import (
	"reflect"
	"unsafe"
)

func BytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func StringToBytes(s string) (data []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return data
}
