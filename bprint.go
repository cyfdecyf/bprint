package main

// The binary format specifier uses the same syntax as Ruby's Array.unpack
//
// c: signed 8-bit integer
// s: signed 16-bit integer
// l: signed 32-bit integer
// q: signed 65-bit integer
//
// Use upper case letter for unsigned integer.
//
// Numbers following the letter means how many times the previous string
// should be repeated.
//
// TODO Also check output format string to make sure number of fields matches

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.1"

func printVersion() {
	fmt.Println("bprint version", version)
	os.Exit(0)
}

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

var (
	printRecordCnt bool
	recordCnt      int
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

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

const noDesc byte = 255

func parseBinaryFmtSpec(binFmt string) (formatDesc []byte) {
	formatDesc = make([]byte, 0)
	var repeatNum int
	var prevDesc byte = noDesc
	for i := 0; i < len(binFmt); i++ {
		desc, ok := descCharMap[binFmt[i]]
		if ok {
			if repeatNum != 0 {
				// The original letter specifier is already added, so minus 1
				for i := 0; i < repeatNum-1; i++ {
					formatDesc = append(formatDesc, prevDesc)
				}
				repeatNum = 0
			}
			formatDesc = append(formatDesc, desc)
			prevDesc = desc
		} else {
			if isDigit(binFmt[i]) {
				if prevDesc == noDesc {
					// Number must follow a previous specifier
					panic("Data specifier error: repeat number without previous data specifier")
				}
				// Parse repeat number
				repeatNum = repeatNum*10 + int(binFmt[i]) - '0'
			} else {
				panic(fmt.Sprintf("Data specifier '%c' not supported", binFmt[i]))
			}
		}
	}
	// If the last specifier is a number
	for i := 0; i < repeatNum-1; i++ {
		formatDesc = append(formatDesc, prevDesc)
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
	if printRecordCnt {
		fmt.Printf("%d: ", recordCnt)
	}
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
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	var binaryFmt, outputFmt string
	var version bool
	flag.StringVar(&binaryFmt, "e", "",
		"binary format specifier. c,s,l,q for signed 8,16,32,64-bit int. Upper case for unsigned int")
	flag.StringVar(&outputFmt, "p", "",
		"printf style output format, size is implicit from binary format specifier")
	flag.BoolVar(&version, "version", false,
		"print version information")
	flag.BoolVar(&printRecordCnt, "c", false,
		"print record count")
	flag.Parse()

	if version {
		printVersion()
	}

	binFilePath := flag.Arg(0)
	if binaryFmt == "" {
		binaryFmt = defautlBinaryFmt
		outputFmt = defaultOutputFmt
	}

	binReader, _ := openFile(binFilePath)

	formatDesc := parseBinaryFmtSpec(binaryFmt)
	formatDescLen := len(formatDesc)
	data := make([]interface{}, formatDescLen, formatDescLen)

	n := 0
	var err error
	for n, err = readData(binReader, formatDesc, data); err == nil; n, err = readData(binReader, formatDesc, data) {
		recordCnt++
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
