package hlsdl

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grafov/m3u8"
	"gopkg.in/cheggaaa/pb.v1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// HlsDl present a HLS downloader
type HlsDl struct {
	client    *http.Client
	headers   map[string]string
	dir       string
	hlsURL    string
	workers   int
	bar       *pb.ProgressBar
	enableBar bool
}

type Segment struct {
	*m3u8.MediaSegment
	Path string
}

type DownloadResult struct {
	Err   error
	SeqId uint64
}

func New(hlsURL string, headers map[string]string, dir string, workers int, enableBar bool) *HlsDl {
	hlsdl := &HlsDl{
		hlsURL:    hlsURL,
		dir:       dir,
		client:    &http.Client{},
		workers:   workers,
		enableBar: enableBar,
		headers:   headers,
	}

	return hlsdl
}

func wait(wg *sync.WaitGroup) chan bool {
	c := make(chan bool, 1)
	go func() {
		wg.Wait()
		c <- true
	}()
	return c
}

func (hlsDl *HlsDl) downloadSegment(segment *Segment) error {
	req, err := newRequest(segment.URI, hlsDl.headers)
	if err != nil {
		return err
	}

	res, err := hlsDl.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}

	file, err := os.Create(segment.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}

	return nil
}

func (hlsDl *HlsDl) downloadSegments(segments []*Segment) error {

	wg := &sync.WaitGroup{}
	wg.Add(hlsDl.workers)

	finishedChan := wait(wg)
	quitChan := make(chan bool)
	segmentChan := make(chan *Segment)
	downloadResultChan := make(chan *DownloadResult, hlsDl.workers)

	for i := 0; i < hlsDl.workers; i++ {
		go func() {
			defer wg.Done()

			for segment := range segmentChan {

				tried := 0
			DOWNLOAD:
				tried++

				select {
				case <-quitChan:
					return
				default:
				}

				if err := hlsDl.downloadSegment(segment); err != nil {
					if strings.Contains(err.Error(), "connection reset by peer") && tried < 3 {
						time.Sleep(time.Second)
						log.Println("Retry download segment ", segment.SeqId)
						goto DOWNLOAD
					}

					downloadResultChan <- &DownloadResult{Err: err, SeqId: segment.SeqId}
					return
				}

				downloadResultChan <- &DownloadResult{SeqId: segment.SeqId}
			}
		}()
	}

	go func() {
		defer close(segmentChan)

		for _, segment := range segments {
			segName := fmt.Sprintf("seg%d.ts", segment.SeqId)
			segment.Path = filepath.Join(hlsDl.dir, segName)

			select {
			case segmentChan <- segment:
			case <-quitChan:
				return
			}
		}

	}()

	if hlsDl.enableBar {
		hlsDl.bar = pb.New(len(segments)).SetMaxWidth(100).Prefix("Downloading...")
		hlsDl.bar.ShowElapsedTime = true
		hlsDl.bar.Start()
	}

	defer func() {
		if hlsDl.enableBar {
			hlsDl.bar.Finish()
		}
	}()

	for {
		select {
		case <-finishedChan:
			return nil
		case result := <-downloadResultChan:
			if result.Err != nil {
				close(quitChan)
				return result.Err
			}

			if hlsDl.enableBar {
				hlsDl.bar.Increment()
			}
		}
	}

}

func (hlsDl *HlsDl) join(dir string, segments []*Segment) (string, error) {
	fmt.Println("Joining segments")

	filepath := filepath.Join(dir, "video.ts")

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})

	for _, segment := range segments {

		d, err := hlsDl.decrypt(segment)
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

	return filepath, nil
}

func (hlsDl *HlsDl) Download() (string, error) {
	segs, err := parseHlsSegments(hlsDl.hlsURL, hlsDl.headers)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(hlsDl.dir, os.ModePerm); err != nil {
		return "", err
	}

	if err := hlsDl.downloadSegments(segs); err != nil {
		return "", err
	}

	filepath, err := hlsDl.join(hlsDl.dir, segs)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
