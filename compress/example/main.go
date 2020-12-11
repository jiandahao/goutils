package main

import (
	"bytes"
	"fmt"

	"github.com/jiandahao/goutils/compress"
)

func main() {

	var buf bytes.Buffer
	buf.Write([]byte("randomdata 123456789"))

	var buffer bytes.Buffer
	if err := compress.Compress(&buf, &buffer); err != nil {
		fmt.Println(err)
		return
	}

	var output bytes.Buffer
	if err := compress.Decompress(&buffer, &output); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output.Bytes()))
}
