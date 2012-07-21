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
	for i := 0; i < b.N; i++ {
		bufReader, readCloser := openFile("testdata/bindata")
		b.StartTimer()
		for err == nil {
			_, err = readData(bufReader, readFunc, data)
		}
		b.StopTimer()
		readCloser.Close()
	}
}
