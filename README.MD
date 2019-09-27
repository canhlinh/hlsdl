# HLS downloader
This is a library to support downloading a hls file. 


### Features:
* Concurrent download segments with multiple http connections
* Decrypt hls encoded segments
* Auto retry download


### Todo:
* Allow adding http headers

### How to use

Get library
```
go get github.com/canhlinh/hlsdl
```

Sample:

```
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

```