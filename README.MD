# HLS downloader
This is a library to support downloading a hls file. 


### Features:
* Concurrent download segments with multiple http connections
* Decrypt hls encoded segments
* Auto retry download


### Todo:
* Allow adding http headers
* Display downloading progress bar

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

### How to build a console application

Build for linux
```
make build-linux
```

Build for windows
```
make build-windows
```

Run the application

```
./bin/hlsdl --help
```

```
./bin/hlsdl -u https://bitdash-a.akamaihd.net/content/sintel/hls/video/1500kbit.m3u8 -d download -w 10
```