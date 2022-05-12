package util

import "unsafe"

func ByteToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}
