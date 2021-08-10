package hlsdl

import (
	"fmt"
	"os"
	"testing"
)

func TestDescrypt(t *testing.T) {
	segs, err := parseHlsSegments("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", nil)
	if err != nil {
		t.Fatal(err)
	}

	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", nil, "./download", 2, false)
	seg := segs[0]
	seg.Path = fmt.Sprintf("%s/seg%d.ts", hlsDl.dir, seg.SeqId)
	if err := hlsDl.downloadSegment(seg); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(seg.Path)

	if _, err := hlsDl.decrypt(seg); err != nil {
		t.Fatal(err)
	}
}

func TestDownload(t *testing.T) {
	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", nil, "./download", 2, false)
	filepath, err := hlsDl.Download()
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(filepath)
}
