package main

// TODO Also check output format string to make sure number of fields matches

import (
	"flag"
	"fmt"
	"os"
	"io"
	"bufio"
	"encoding/binary"
)

var byteOrder = binary.LittleEndian

var (
	binaryFmt string
	outputFmt string
	binReader io.Reader
)

type intReadFunc func(io.Reader) (interface{}, error)

func readI8(reader io.Reader) (interface{}, error) {
	var i int8
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readI16(reader io.Reader) (interface{}, error) {
	var i int16
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readI32(reader io.Reader) (interface{}, error) {
	var i int32
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readI64(reader io.Reader) (interface{}, error) {
	var i int64
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readU8(reader io.Reader) (interface{}, error) {
	var i uint8
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readU16(reader io.Reader) (interface{}, error) {
	var i uint16
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readU32(reader io.Reader) (interface{}, error) {
	var i uint32
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

func readU64(reader io.Reader) (interface{}, error) {
	var i uint64
	err := binary.Read(reader, byteOrder, &i)
	return i, err
}

var readFuncMap = map[byte]intReadFunc{
	'c': readI8,
	's': readI16,
	'l': readI32,
	'q': readI64,

	'C': readU8,
	'S': readU16,
	'L': readU32,
	'Q': readU64,
}

func parseBinaryFormatStr(binFmt string) (binSpec []intReadFunc) {
	binSpec = make([]intReadFunc, 0)
	for i := 0; i < len(binFmt); i++ {
		f, ok := readFuncMap[binFmt[i]]
		if !ok {
			fmt.Printf("Data specifier '%c' not supported\n", binFmt[i])
			os.Exit(1)
		}
		binSpec = append(binSpec, f)
	}
	return
}

// Read binary data
func readData(readFuncs []intReadFunc, data []interface{}) (n int, err error) {
	for i, rf := range readFuncs {
		data[i], err = rf(binReader)

		if err != nil {
			break
		}
		n++
	}
	return
}

func printData(data []interface{}) {
	fmt.Printf(outputFmt, data...)
	fmt.Println()
}

func init() {
	flag.StringVar(&binaryFmt, "e", "",
		"Binary format specifier. c,s,l,q for 8,16,32,64 bit signed int. Upper case for unsigned int.")
	flag.StringVar(&outputFmt, "p", "",
		"Printf style format, size is implicit from format specifier.")
}

func main() {
	flag.Parse()

	binFile := flag.Arg(0)
	if binaryFmt == "" {
		binaryFmt = "CCCCCCCCCCCCCCCC"
		outputFmt = "%02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x"
	}

	var err error
	if binFile == "" {
		binReader = os.Stdin
	} else {
		binReader, err = os.Open(binFile)
		if err != nil {
			fmt.Println("While opening file:", err)
			os.Exit(1)
		}
	}
	binReader = bufio.NewReader(binReader)

	binSpec := parseBinaryFormatStr(binaryFmt)
	binSpecLen := len(binSpec)
	data := make([]interface{}, binSpecLen, binSpecLen)
	var n int
	for n, err = readData(binSpec, data); err == nil; n, err = readData(binSpec, data) {
		printData(data)
	}
	// Not enough data for the final line, print out what have been read
	// if n != 0 && n != readFuncLen {
	if n != 0 {
		printData(data[0:n])
	}
	if err != io.EOF {
		if err == io.ErrUnexpectedEOF {
			fmt.Println("EOF: final data not enough for the last field")
		} else {
			fmt.Println("While reading data:", err)
		}
	}
}
