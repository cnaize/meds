package util

import "unsafe"

func BytesToString(str []byte) string {
	return unsafe.String(unsafe.SliceData(str), len(str))
}

func StringToBytes(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}
