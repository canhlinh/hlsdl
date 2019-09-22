package hlsdl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/grafov/m3u8"
)

// HlsDl present a HLS downloader
type HlsDl struct {
	client  *http.Client
	dir     string
	hlsURL  string
	workers int
}

type Segment struct {
	*m3u8.MediaSegment
	Path string
}

type DownloadError struct {
	Err error
}

func New(hlsURL string, dir string, workers int) *HlsDl {
	return &HlsDl{
		hlsURL:  hlsURL,
		dir:     dir,
		client:  &http.Client{},
		workers: workers,
	}
}

func wait(wg *sync.WaitGroup) chan bool {
	c := make(chan bool)
	go func() {
		wg.Wait()
		c <- true
	}()
	return c
}

func (hlsDl *HlsDl) downloadSegment(segment *Segment) error {
	fmt.Printf("Downloading segment %d \n", segment.SeqId)

	res, err := hlsDl.client.Get(segment.URI)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}

	file, err := os.Create(segment.Path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}

	fmt.Printf("Downloaded segment %d \n", segment.SeqId)
	return nil
}

func (hlsDl *HlsDl) downloadMediaSegments(segments []*Segment) error {

	wg := &sync.WaitGroup{}
	wg.Add(hlsDl.workers)

	waitChan := wait(wg)
	segmentChan := make(chan *Segment)
	errHandlerChan := make(chan *DownloadError, hlsDl.workers)

	for i := 0; i < hlsDl.workers; i++ {
		go func() {
			defer wg.Done()

			for segment := range segmentChan {
				if err := hlsDl.downloadSegment(segment); err != nil {
					errHandlerChan <- &DownloadError{err}
				}
			}
		}()
	}

	go func() {
		for _, segment := range segments {
			segment.Path = fmt.Sprintf("%s/seg%d.ts", hlsDl.dir, segment.SeqId)
			segmentChan <- segment
		}

		fmt.Println("Closing segment chan")
		close(segmentChan)
	}()

	for {
		select {
		case <-waitChan:
			fmt.Println("Closing wait chan")
			return nil
		case err := <-errHandlerChan:
			close(segmentChan)
			return err.Err
		}
	}
}

func (hlsDl *HlsDl) join(dir string, segments []*Segment) (string, error) {
	filepath := fmt.Sprintf("%s/video.ts", dir)

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})

	for _, segment := range segments {

		d, err := Decrypt(segment)
		if err != nil {
			return "", err
		}

		if _, err := file.Write(d); err != nil {
			return "", err
		}

		if err := os.RemoveAll(segment.Path); err != nil {
			return "", err
		}
	}

	file.Close()
	return filepath, nil
}

func (hlsDl *HlsDl) Download() error {
	segs, err := parseHlsSegments(hlsDl.hlsURL)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(hlsDl.dir, os.ModePerm); err != nil {
		return err
	}

	if err := hlsDl.downloadMediaSegments(segs); err != nil {
		return err
	}

	filepath, err := hlsDl.join(hlsDl.dir, segs)
	if err != nil {
		return err
	}

	fmt.Println(filepath)

	return nil
}
