package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/OneOfOne/xxhash"
)

func main() {
	h := xxhash.New64()
	// r, err := os.Open("......")
	// defer f.Close()
	r := strings.NewReader("hello")
	io.Copy(h, r)
	fmt.Println("xxhash.Backend:", xxhash.Backend)
	fmt.Println("File checksum:", h.Sum64())
}
//testing