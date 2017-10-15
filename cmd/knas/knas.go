package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/karlek/skillnad"
)

var amount int
var out string

func init() {
	flag.IntVar(&amount, "a", 10, "amount of glitched bytes.")
	flag.StringVar(&out, "o", "", "output filename.")

	flag.Usage = usage
	rand.Seed(time.Now().UnixNano())
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
	}
	for _, filename := range flag.Args() {
		err := jpg(filename)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func jpg(filename string) (err error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	glitchJpg(buf, headerSize(buf), amount)
	if out == "" {
		out = skillnad.RemoveExt(filename) + "-knas.jpg"
	}
	return ioutil.WriteFile(out, buf, 0700)
}

func glitchByte(buf *[]byte, headerLen int) {
	body := len(*buf) - headerLen - 4
	// Change a random byte in the jpg, except in the header.
	index := headerLen + int(rand.Float64()*float64(body))
	(*buf)[index] = byte(rand.Float64() * 256)
}

func glitchJpg(buf []byte, headerLen, amount int) {
	for i := 0; i < amount; i++ {
		glitchByte(&buf, headerLen)
	}
}

func headerSize(buf []byte) int {
	// index := bytes.Index(buf, []byte{0xff, 0xda})
	// if index == -1 {
	// 	log.Fatalln("index == -1")
	// }
	// return index
	headerLen := 417
	l := len(buf)
	for i := 0; i < l; i++ {
		if buf[i] == 0xff && buf[i+1] == 0xda {
			headerLen += 2
		}
	}
	return headerLen
}
