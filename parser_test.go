package hlsdl

import (
	"testing"
)

func Test(t *testing.T) {
	_, err := parseHlsSegments("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8")
	if err != nil {
		t.Fatal(err)
	}
}
