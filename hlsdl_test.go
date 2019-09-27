package hlsdl

import (
	"fmt"
	"os"
	"testing"
)

func TestDescrypt(t *testing.T) {
	segs, err := parseHlsSegments("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8")
	if err != nil {
		t.Fatal(err)
	}

	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", "./download", 2)

	seg := segs[0]
	seg.Path = fmt.Sprintf("%s/seg%d.ts", hlsDl.dir, seg.SeqId)
	if err := hlsDl.downloadSegment(seg); err != nil {
		t.Fatal(err)
	}

	if _, err := hlsDl.Decrypt(seg); err != nil {
		t.Fatal(err)
	}
}

func TestDownload(t *testing.T) {

	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", "./download", 2)
	filepath, err := hlsDl.Download()
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(filepath)
}
