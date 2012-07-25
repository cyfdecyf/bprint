package main

// The binary format string uses the same syntax as Ruby's Array.unpack
//
// c: signed 8-bit integer
// s: signed 16-bit integer
// l: signed 32-bit integer
// q: signed 64-bit integer
//
// Use upper case letter for unsigned integer.
//
// Numbers following the letter means how many times the previous string
// should be repeated.

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const version = "0.2.1"

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

type intType byte

type intDesc struct {
	typeId intType
	size   int
}

const noType intType = 255

const (
	I8 intType = iota
	I16
	I32
	I64

	U8
	U16
	U32
	U64
)

var descCharMap = map[byte]intDesc{
	'c': {I8, 1},
	's': {I16, 2},
	'l': {I32, 4},
	'q': {I64, 8},

	'C': {U8, 1},
	'S': {U16, 2},
	'L': {U32, 4},
	'Q': {U64, 8},
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func parseBinaryFmt(binFmt string) (formatField []intType, recSize int) {
	formatField = make([]intType, 0)
	var repeatNum int
	prevDesc := intDesc{noType, -1}
	for i := 0; i < len(binFmt); i++ {
		desc, ok := descCharMap[binFmt[i]]
		if ok {
			if repeatNum != 0 {
				// The original field is already added, so minus 1
				for i := 0; i < repeatNum-1; i++ {
					formatField = append(formatField, prevDesc.typeId)
				}
				recSize += (repeatNum - 1) * prevDesc.size
				repeatNum = 0
			}
			formatField = append(formatField, desc.typeId)
			prevDesc = desc
			recSize += desc.size
		} else {
			if isDigit(binFmt[i]) {
				if prevDesc.typeId == noType {
					// Number must follow a previous field
					panic("Data field error: repeat number without previous data field")
				}
				// Parse repeat number
				repeatNum = repeatNum*10 + int(binFmt[i]) - '0'
			} else {
				panic(fmt.Sprintf("Data field '%c' not supported", binFmt[i]))
			}
		}
	}
	// If the last field has repeat cnt
	for i := 0; i < repeatNum-1; i++ {
		formatField = append(formatField, prevDesc.typeId)
	}
	if repeatNum != 0 {
		recSize += (repeatNum - 1) * prevDesc.size
	}
	return
}

func readData(binReader io.Reader, formatField []intType, data []interface{}) (n int, err error) {
	for i, v := range formatField {
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

var (
	recordCnt  int
	recordSize int
	offSet     int
)

const offsetFmt = "%07x "

func printData(printFmt string, data []interface{}) {
	if opt.printOffset {
		fmt.Printf(offsetFmt, offSet)
	}
	if opt.printRecordCnt {
		fmt.Printf("%d: ", recordCnt)
	}
	fmt.Printf(printFmt, data...)
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

func repeatWithSep(rep, sep string, cnt int) string {
	printFmt := strings.Repeat(rep+sep, cnt)
	return printFmt[:len(printFmt)-len(sep)]
}

func generatePrintFmt(cnt int, sep string) string {
	return repeatWithSep("%02x", sep, cnt)
}

func processPrintFmt(printFmt string) string {
	// Format like "%02d[sep]8#", "%d" will be repeated 8 times, with
	// seperator inserted. The # is used to mark the end of separator and repeat count,
	// it's not necessary, only to make it easier to see where is the end of the field.
	printFieldPat, err := regexp.Compile("(%[^cdxo%]*[cdxo])([^\\d]*)(\\d+)#")
	if err != nil {
		panic(err)
	}
	mat := printFieldPat.FindAllStringSubmatchIndex(printFmt, -1)
	if mat == nil {
		return printFmt
	}

	buf := new(bytes.Buffer)
	prevIdx := 0
	for _, v := range mat {
		buf.WriteString(printFmt[prevIdx:v[0]])
		prevIdx = v[1]
		if v[0] > 0 && printFmt[v[0]-1] == '%' {
			// Do not parse field following %%
			buf.WriteString(printFmt[v[0]:v[1]])
			continue
		}

		field := printFmt[v[2]:v[3]]
		sep := printFmt[v[4]:v[5]]
		cntStr := printFmt[v[6]:v[7]]
		if sep == "" {
			sep = " "
		}
		cnt, err := strconv.Atoi(cntStr)
		if err != nil {
			panic(err)
		}

		buf.WriteString(repeatWithSep(field, sep, cnt))
	}
	buf.WriteString(printFmt[prevIdx:])

	return buf.String()
}

func countPrintFmtField(printFmt string) int {
	fieldStr := "%[^cdxo%]*[cdxo]"
	// fieldStr must have a non-% preceeding or start from the beginning of line
	printFieldPat, err := regexp.Compile("([^%]{1}" + fieldStr + "|^" + fieldStr + ")")
	if err != nil {
		panic(err)
	}

	return len(printFieldPat.FindAllStringIndex(printFmt, -1))
}

func readOptionFromFile() {
	f, err := os.Open(opt.formatFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	const fmtCnt = 2
	line := [fmtCnt]string{}
	for i := 0; i < fmtCnt; i++ {
		line[i], err = buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(fmt.Sprintf("Error reading from format file %v", err))
		}
	}
	if opt.binaryFmt == "" {
		opt.binaryFmt = strings.TrimRight(line[0], "\n")
	}
	if opt.printFmt == "" {
		opt.printFmt = strings.TrimRight(line[1], "\n")
	}
}

const (
	defautlBinaryFmt = "C16"
)

var opt struct {
	printRecordCnt bool
	printOffset    bool
	printVersion   bool
	binaryFmt      string
	printFmt       string
	formatFile     string
}

func init() {
	flag.StringVar(&opt.binaryFmt, "e", "",
		"binary format string. c,s,l,q for signed 8,16,32,64-bit int. Upper case for unsigned int")
	flag.StringVar(&opt.printFmt, "p", "",
		"printf style format string, size is implicit from binary format string, default to %02x for each field")
	flag.StringVar(&opt.formatFile, "f", "",
		"read binary and print format from file. 1st line for binary format, 2nd line for print format (optional)\n\t "+
			"command line option overrides option in file")
	flag.BoolVar(&opt.printVersion, "version", false,
		"print version information")
	flag.BoolVar(&opt.printRecordCnt, "c", false,
		"print record count")
	flag.BoolVar(&opt.printOffset, "o", false,
		"print record count")
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	flag.Parse()
	if opt.printVersion {
		printVersion()
	}
	if opt.formatFile != "" {
		readOptionFromFile()
	}
	if opt.binaryFmt == "" {
		opt.binaryFmt = defautlBinaryFmt
	}
	formatField, recordSize := parseBinaryFmt(opt.binaryFmt)
	formatFieldCnt := len(formatField)
	if opt.printFmt == "" {
		opt.printFmt = generatePrintFmt(formatFieldCnt, " ")
	} else {
		opt.printFmt = processPrintFmt(opt.printFmt)
	}
	// Check if binary and print format has the same field count
	printFieldCnt := countPrintFmtField(opt.printFmt)
	if printFieldCnt != formatFieldCnt {
		panic(fmt.Sprintf("Binary format has %d fields, print fmt has %d fields. Not match.",
			formatFieldCnt, printFieldCnt))
	}

	opt.printFmt += "\n"

	binFilePath := flag.Arg(0)
	binReader, f := openFile(binFilePath)
	defer f.Close()

	data := make([]interface{}, formatFieldCnt, formatFieldCnt)
	n := 0
	var err error
	for n, err = readData(binReader, formatField, data); err == nil; n, err = readData(binReader, formatField, data) {
		recordCnt++
		printData(opt.printFmt, data)
		offSet += recordSize
	}
	// Not enough data for the final line, print out what have been read
	if n != 0 {
		printData(opt.printFmt, data[:n])
	} else if opt.printOffset {
		fmt.Printf(offsetFmt+"\n", offSet)
	}
	if err != io.EOF {
		if err == io.ErrUnexpectedEOF {
			fmt.Println("EOF: final data not enough for the last field")
		} else {
			fmt.Println("While reading data:", err)
		}
	}
}
