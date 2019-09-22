package main

import (
	"github.com/canhlinh/hlsdl"
)

func main() {
	hlsDL := hlsdl.New("https://bitdash-a.akamaihd.net/content/sintel/hls/video/10000kbit.m3u8", "download", 10)
	if err := hlsDL.Download(); err != nil {
		panic(err)
	}
}
