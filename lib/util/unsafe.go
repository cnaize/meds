package util

import "unsafe"

func BytesToString(str []byte) string {
	data := unsafe.SliceData(str)
	return unsafe.String(data, len(str))
}

func StringToBytes(str string) []byte {
	data := unsafe.StringData(str)
	return unsafe.Slice(data, len(str))
}
