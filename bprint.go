package main

// TODO Also check output format string to make sure number of fields matches

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

var byteOrder = binary.LittleEndian

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

func parseBinaryFormatStr(binFmt string) (readFunc []intReadFunc) {
	readFunc = make([]intReadFunc, 0)
	for i := 0; i < len(binFmt); i++ {
		f, ok := readFuncMap[binFmt[i]]
		if !ok {
			fmt.Printf("Data specifier '%c' not supported\n", binFmt[i])
			os.Exit(1)
		}
		readFunc = append(readFunc, f)
	}
	return
}

// Read binary data
func readData(binReader io.Reader, readFuncs []intReadFunc, data []interface{}) (n int, err error) {
	for i, rf := range readFuncs {
		data[i], err = rf(binReader)

		if err != nil {
			break
		}
		n++
	}
	return
}

func printData(outputFmt string, data []interface{}) {
	fmt.Printf(outputFmt, data...)
	fmt.Println()
}

func openFile(path string) (reader io.Reader, ioReader io.ReadCloser) {
	if path == "" {
		ioReader = os.Stdin
	} else {
		var err error
		ioReader, err = os.Open(path)
		if err != nil {
			fmt.Println("While opening file:", err)
			os.Exit(1)
		}
	}
	reader = bufio.NewReader(ioReader)
	return
}

const (
	defautlBinaryFmt = "CCCCCCCCCCCCCCCC"
	defaultOutputFmt = "%02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x %02x"
)

func main() {
	var binaryFmt, outputFmt string
	flag.StringVar(&binaryFmt, "e", "",
		"Binary format specifier. c,s,l,q for 8,16,32,64 bit signed int. Upper case for unsigned int.")
	flag.StringVar(&outputFmt, "p", "",
		"Printf style format, size is implicit from format specifier.")
	flag.Parse()

	binFilePath := flag.Arg(0)
	if binaryFmt == "" {
		binaryFmt = defautlBinaryFmt
		outputFmt = defaultOutputFmt
	}

	binReader, _ := openFile(binFilePath)

	readFunc := parseBinaryFormatStr(binaryFmt)
	readFuncLen := len(readFunc)
	data := make([]interface{}, readFuncLen, readFuncLen)
	var n int
	var err error
	for n, err = readData(binReader, readFunc, data); err == nil; n, err = readData(binReader, readFunc, data) {
		printData(outputFmt, data)
	}
	// Not enough data for the final line, print out what have been read
	// if n != 0 && n != readFuncLen {
	if n != 0 {
		printData(outputFmt, data[0:n])
	}
	if err != io.EOF {
		if err == io.ErrUnexpectedEOF {
			fmt.Println("EOF: final data not enough for the last field")
		} else {
			fmt.Println("While reading data:", err)
		}
	}
}
