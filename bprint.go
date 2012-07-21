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

var (
	i8  int8
	i16 int16
	i32 int32
	i64 int64

	u8  uint8
	u16 uint16
	u32 uint32
	u64 uint64
)

const (
	I8 byte = iota
	I16
	I32
	I64
	U8
	U16
	U32
	U64
)

var descCharMap = map[byte]byte{
	'c': I8,
	's': I16,
	'l': I32,
	'q': I64,

	'C': U8,
	'S': U16,
	'L': U32,
	'Q': U64,
}

func parseBinaryFormatStr(binFmt string) (formatDesc []byte) {
	formatDesc = make([]byte, 0)
	for i := 0; i < len(binFmt); i++ {
		desc, ok := descCharMap[binFmt[i]]
		if !ok {
			fmt.Printf("Data specifier '%c' not supported\n", binFmt[i])
			os.Exit(1)
		}
		formatDesc = append(formatDesc, desc)
	}
	return
}

func readData(binReader io.Reader, formatDesc []byte, data []interface{}) (n int, err error) {
	for i, v := range formatDesc {
		switch v {
		case I8:
			err = binary.Read(binReader, byteOrder, &i8)
			data[i] = i8
		case I16:
			err = binary.Read(binReader, byteOrder, &i16)
			data[i] = i16
		case I32:
			err = binary.Read(binReader, byteOrder, &i32)
			data[i] = i32
		case I64:
			err = binary.Read(binReader, byteOrder, &i64)
			data[i] = i64

		case U8:
			err = binary.Read(binReader, byteOrder, &u8)
			data[i] = u8
		case U16:
			err = binary.Read(binReader, byteOrder, &u16)
			data[i] = u16
		case U32:
			err = binary.Read(binReader, byteOrder, &u32)
			data[i] = u32
		case U64:
			err = binary.Read(binReader, byteOrder, &u64)
			data[i] = u64
		}

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

	formatDesc := parseBinaryFormatStr(binaryFmt)
	formatDescLen := len(formatDesc)
	data := make([]interface{}, formatDescLen, formatDescLen)
	var n int
	var err error
	for n, err = readData(binReader, formatDesc, data); err == nil; n, err = readData(binReader, formatDesc, data) {
		printData(outputFmt, data)
	}
	// Not enough data for the final line, print out what have been read
	// if n != 0 && n != formatDescLen {
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
