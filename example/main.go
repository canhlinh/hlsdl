package main

import (
	"fmt"

	"github.com/canhlinh/hlsdl"
)

func main() {
	hlsDL := hlsdl.New("https://bitdash-a.akamaihd.net/content/sintel/hls/video/1500kbit.m3u8", "download", 64)
	filepath, err := hlsDL.Download()
	if err != nil {
		panic(err)
	}

	fmt.Println(filepath)
}
