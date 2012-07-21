package main

import (
	"testing"
)

func BenchmarkReadData(b *testing.B) {
	// Benchmark setup:
	//   Intel Q6600 CPU, Debian 6 with Go 1.0.1
	//   11MB random binary data, with default binary format specifier
	//
	// Execution time:
	//   use function         ~3.6s
	//   use switch statement ~2.7s
	b.StopTimer()
	formatDesc := parseBinaryFormatStr(defautlBinaryFmt)
	formatDescLen := len(formatDesc)
	data := make([]interface{}, formatDescLen, formatDescLen)

	var err error
	for i := 0; i < b.N; i++ {
		bufReader, readCloser := openFile("testdata/bindata")
		b.StartTimer()
		for err == nil {
			_, err = readData(bufReader, formatDesc, data)
		}
		b.StopTimer()
		readCloser.Close()
	}
}
