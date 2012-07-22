package main

import (
	"bytes"
	"testing"
)

type binFmtData struct {
	binFmt  string
	fmtDesc []byte
}

func TestParseBinaryFmtSpec(t *testing.T) {
	testData := []binFmtData{
		{"cslqCSLQ", []byte{I8, I16, I32, I64, U8, U16, U32, U64}},
		{"c2", []byte{I8, I8}},
		{"s1q", []byte{I16, I64}},
		{"c11q2", []byte{I8, I8, I8, I8, I8, I8, I8, I8, I8, I8, I8, I64, I64}},
	}

	for _, td := range testData {
		res := parseBinaryFmtSpec(td.binFmt)
		if bytes.Compare(td.fmtDesc, res) != 0 {
			t.Error("binary fmt spec:", td.binFmt, "not parsed correctly, got", res)
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
	formatDesc := parseBinaryFmtSpec(defautlBinaryFmt)
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
