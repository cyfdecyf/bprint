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

var (
	binaryFmt string
	outputFmt string
	binReader io.Reader
)

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

var specifierSet = map[byte]bool{
	'c': true,
	'C': true,
	's': true,
	'S': true,
	'l': true,
	'L': true,
	'q': true,
	'Q': true,
}

func parseBinaryFormatStr(binFmt string) (binSpec []byte) {
	binSpec = make([]byte, 0)
	for i := 0; i < len(binFmt); i++ {
		_, ok := specifierSet[binFmt[i]]
		if !ok {
			fmt.Printf("Data specifier '%c' not supported\n", binFmt[i])
			os.Exit(1)
		}
		binSpec = append(binSpec, binFmt[i])
	}
	return
}

// Read binary data according to the format descriptor
func readData(desc []byte, data []interface{}) (n int, err error) {
	for i, v := range desc {
		switch v {
		case 'c':
			err = binary.Read(binReader, binary.LittleEndian, &i8)
			data[i] = i8
		case 'C':
			err = binary.Read(binReader, binary.LittleEndian, &u8)
			data[i] = u8
		case 's':
			err = binary.Read(binReader, binary.LittleEndian, &i16)
			data[i] = i16
		case 'S':
			err = binary.Read(binReader, binary.LittleEndian, &u16)
			data[i] = u16
		case 'l':
			err = binary.Read(binReader, binary.LittleEndian, &i32)
			data[i] = i32
		case 'L':
			err = binary.Read(binReader, binary.LittleEndian, &u32)
			data[i] = u32
		case 'q':
			err = binary.Read(binReader, binary.LittleEndian, &i64)
			data[i] = i64
		case 'Q':
			err = binary.Read(binReader, binary.LittleEndian, &u64)
			data[i] = u64
		}
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				fmt.Println("ERROR: not enough data for the next field")
				os.Exit(1)
			} else if err != io.EOF {
				fmt.Println("While reading data:", err)
			}
			return
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
	if err == io.EOF && n != 0 && n != binSpecLen {
		// fill 0 for last record's field that's not filled
		// fmt.Println("fill last and print")
		for i := 0; i < binSpecLen-n; i++ {
			data[n+i] = byte(0)
		}
		printData(data)
	}
}
