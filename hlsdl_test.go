package hlsdl

import (
	"fmt"
	"testing"
)

func TestDownloadEncrypted(t *testing.T) {
	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny_encrypted/stream-800/index.m3u8", nil, "./download", "", 2, false)
	filepath, err := hlsDl.Download()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(filepath)
	// os.RemoveAll(filepath)
}

func TestDownloadNormal(t *testing.T) {
	hlsDl := New("https://cdn.theoplayer.com/video/big_buck_bunny/stream-800/index.m3u8", nil, "./download", "", 2, false)
	filepath, err := hlsDl.Download()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(filepath)
	// os.RemoveAll(filepath)
}
