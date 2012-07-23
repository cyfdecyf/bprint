package main

import (
	"testing"
)

type binFmtData struct {
	binFmt  string
	fmtDesc []intType
	size    int
}

func TestParseBinaryFmtSpec(t *testing.T) {
	testData := []binFmtData{
		{"cslqCSLQ", []intType{I8, I16, I32, I64, U8, U16, U32, U64}, 30},
		{"c2", []intType{I8, I8}, 2},
		{"s1q", []intType{I16, I64}, 10},
		{"c11q2", []intType{I8, I8, I8, I8, I8, I8, I8, I8, I8, I8, I8, I64, I64}, 27},
	}

	for _, td := range testData {
		res, size := parseBinaryFmtSpec(td.binFmt)
		for i, v := range res {
			if td.fmtDesc[i] != v {
				t.Error("binary fmt spec:", td.binFmt, "not parsed correctly, got", res)
			}
		}
		if size != td.size {
			t.Error("binary fmt spec:", td.binFmt, "size should be", td.size,
				", got", size)
		}
	}
}

func TestParseUnsupportedBinaryFormatSpec(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	parseBinaryFmtSpec("ccid")
	t.Error("Should panic for unsuppored specifier")
}

func TestParseNospecNumberBinaryFormatSpec(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	parseBinaryFmtSpec("11clsq")
	t.Error("Should panic for repeat number without spec")
}

func BenchmarkReadData(b *testing.B) {
	// Benchmark setup:
	//   Intel Q6600 CPU, Debian 6 with Go 1.0.1
	//   11MB random binary data, with default binary format specifier
	//
	// Execution time:
	//   use function         ~3.6s
	//   use switch statement ~2.7s
	b.StopTimer()
	formatDesc, _ := parseBinaryFmtSpec(defautlBinaryFmt)
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
