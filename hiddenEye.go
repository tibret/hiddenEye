package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const JPEG_MARKER byte = 0xFF
const COMMENT byte = 0xFE

var START_OF_IMAGE []byte = []byte{JPEG_MARKER, 0xD8}
var START_OF_FRAME_BASELINE []byte = []byte{JPEG_MARKER, 0xC0}
var START_OF_FRAME_PROGRESSIVE []byte = []byte{JPEG_MARKER, 0xC2}
var DEFINE_HUFFMAN_TABLE []byte = []byte{JPEG_MARKER, 0xC4}
var DEFINE_QUANTIZATION_TABLES []byte = []byte{JPEG_MARKER, 0xDB}
var DEFINE_RESTART_INTERVAL []byte = []byte{JPEG_MARKER, 0xDD}
var START_OF_SCAN []byte = []byte{JPEG_MARKER, 0xDA}
var RESTART []byte = []byte{JPEG_MARKER, 0xD0}
var APPLICATION_SPECIFIC []byte = []byte{JPEG_MARKER, 0xE0}
var COMMENT_MARKER []byte = []byte{JPEG_MARKER, 0xFE}
var END_OF_IMAGE []byte = []byte{JPEG_MARKER, 0xD9}

func main() {
	encodePtr := flag.Bool("e", false, "Encode mode")
	decodePtr := flag.Bool("d", false, "Decode mode")

	fileNamePtr := flag.String("file", "", "The file to operate on, must be a jpeg")
	messagePtr := flag.String("message", "", "The message to encode")

	flag.Parse()

	valid := true

	// do validtion
	if *encodePtr && *fileNamePtr == "" {
		fmt.Println("You must supply -file when encoding")
		valid = false
	}

	if *encodePtr && *messagePtr == "" {
		fmt.Println("You must supply -message when encodign")
		valid = false
	}

	if *decodePtr && *fileNamePtr == "" {
		fmt.Println("You must supply -file when decoding")
		valid = false
	}

	if *decodePtr && *encodePtr {
		fmt.Println("You can only specify one mode at a time")
		valid = false
	}

	if !valid {
		return
	}

	if *encodePtr {
		fmt.Println("Encoding message " + *messagePtr + " in file " + *fileNamePtr)
		Encode(*messagePtr, *fileNamePtr)
	}

	if *decodePtr {
		fmt.Println("Decoding file", *fileNamePtr)
		Decode(*fileNamePtr)
	}

}

func Encode(message string, file string) {
	f, err := os.OpenFile(file, os.O_RDWR, 0600)
	check(err)

	//calculate length of message in bytes
	var messageLength = uint16(len(message) + 2)

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, messageLength)

	messageBytes := []byte(message)

	msg := COMMENT_MARKER
	msg = append(msg, lengthBytes...)
	msg = append(msg, messageBytes...)

	bytesWritten, err := f.WriteAt(msg, 2)
	check(err)
	fmt.Printf("Bytes written %d", bytesWritten)
}

func Details(file string) {
	f, err := os.Open(file)
	check(err)

	buff := make([]byte, 32)
	markerBuff := make([]byte, 2)
	bytes_read, err := f.Read(buff)
	check(err)
	offset := 32
	for bytes_read == 32 {
		// fmt.Printf("%d bytes read: %s\n", bytes_read, string(buff[:bytes_read]))
		// fmt.Print("Current bytes : ")
		// for _, curByte := range buff {
		// 	fmt.Printf("%d, ", curByte)
		// }
		// fmt.Print("\n")

		for idx, curByte := range buff {
			if curByte == JPEG_MARKER {
				jump := int64(idx - 32)
				f.Seek(jump, io.SeekCurrent)
				_, err = f.Read(markerBuff)
				check(err)
				PrintMarkerType(markerBuff)

				f.Seek((-jump)-2, io.SeekCurrent)
			}
		}

		// set up the next loop
		bytes_read, err = f.Read(buff)
		check(err)
		offset += 32
	}
}

func Decode(file string) {
	f, err := os.Open(file)
	check(err)

	buff := make([]byte, 32)
	markerBuff := make([]byte, 2)
	bytes_read, err := f.Read(buff)
	check(err)
	offset := 32
	for bytes_read == 32 {
		for idx, curByte := range buff {
			if curByte == JPEG_MARKER {
				jump := int64(idx - 32)
				f.Seek(jump, io.SeekCurrent)
				_, err = f.Read(markerBuff)
				check(err)
				if markerBuff[1] == COMMENT {
					ProcessComment(f)
					return
				}
				f.Seek((-jump)-2, io.SeekCurrent)
			}
		}

		// set up the next loop
		bytes_read, err = f.Read(buff)
		check(err)
		offset += 32
	}

}

func ProcessComment(f *os.File) {
	//if we read the first 2 bytes, it should be the message size in big endian format
	messageLengthBytes := make([]byte, 2)
	_, err := f.Read(messageLengthBytes)
	check(err)
	messageLength := binary.BigEndian.Uint16(messageLengthBytes) - 2
	message := make([]byte, messageLength)
	_, err = f.Read(message)
	fmt.Printf("The secret message is: %s", string(message))
}

func PrintMarkerType(buff []byte) {
	if bytes.Equal(buff, START_OF_IMAGE) {
		fmt.Println("Start of Image")
	} else if bytes.Equal(buff, START_OF_FRAME_BASELINE) {
		fmt.Println("Start of Frame")
	} else if bytes.Equal(buff, START_OF_FRAME_PROGRESSIVE) {
		fmt.Println("Start of Frame")
	} else if bytes.Equal(buff, DEFINE_HUFFMAN_TABLE) {
		fmt.Println("Define huffman table")
	} else if bytes.Equal(buff, DEFINE_QUANTIZATION_TABLES) {
		fmt.Println("Define Quantization table")
	} else if bytes.Equal(buff, DEFINE_RESTART_INTERVAL) {
		fmt.Println("Define Restart Interval")
	} else if bytes.Equal(buff, START_OF_SCAN) {
		fmt.Println("Start of Scan")
	} else if buff[1] >= 208 && buff[1] <= 215 {
		fmt.Println("Restart")
	} else if buff[1] >= 224 && buff[1] <= 239 {
		fmt.Println("Application Specific")
	} else if bytes.Equal(buff, COMMENT_MARKER) {
		fmt.Println("Comment")
	} else if bytes.Equal(buff, END_OF_IMAGE) {
		fmt.Println("End of Image")
	}
}
