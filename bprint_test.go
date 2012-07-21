package main

import (
	"testing"
)

func BenchmarkReadData(b *testing.B) {
	b.StopTimer()
	readFunc := parseBinaryFormatStr(defautlBinaryFmt)
	readFuncLen := len(readFunc)
	data := make([]interface{}, readFuncLen, readFuncLen)

	var err error
	reader := openFile("testdata/bindata")
	b.StartTimer()
	for err == nil {
		_, err = readData(reader, readFunc, data)
	}
}
