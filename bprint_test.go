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

func TestGenerateOutputFormat(t *testing.T) {
	var s string
	s = generatePrintFmt(2, " ")

	if s != "%02x %02x" {
		t.Error("length 2 space sep error")
	}
}

func TestProcessPrintFmt(t *testing.T) {
	td := []string{
		"hello %02d2# %#07x nihao %09o, 2#", "hello %02d %02d %#07x nihao %09o, %09o",
		"%#08c %d %x hello", "%#08c %d %x hello",
		"%#01x1# this %2d,2# world", "%#01x this %2d,%2d world",
		"head %%02d2# end", "head %%02d2# end",
	}

	for i := 0; i < len(td)/2; i++ {
		res := processPrintFmt(td[2*i])
		if res != td[2*i+1] {
			t.Error("Output format processing wrong ", td[2*i], " converted to:", res)
		}
	}
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
