package main

import (
	"fmt"

	"github.com/canhlinh/hlsdl"
)

func main() {
	hlsDL := hlsdl.New("https://playertest.longtailvideo.com/adaptive/oceans_aes/oceans_aes-audio=65000-video=236000.m3u8", "download", 16, true)
	filepath, err := hlsDL.Download()
	if err != nil {
		panic(err)
	}

	fmt.Println(filepath)
}
