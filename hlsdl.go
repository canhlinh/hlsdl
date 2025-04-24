package hlsdl

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/grafov/m3u8"
	"gopkg.in/cheggaaa/pb.v1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// HlsDl present a HLS downloader
type HlsDl struct {
	client     *resty.Client
	headers    map[string]string
	dir        string
	hlsURL     string
	workers    int
	bar        *pb.ProgressBar
	enableBar  bool
	filename   string
	startTime  int64
	segTotal   int64
	segCurrent int64
}

type Segment struct {
	*m3u8.MediaSegment
	Path string
}

type DownloadResult struct {
	Err   error
	SeqId uint64
}

func New(hlsURL string, headers map[string]string, dir, filename string, workers int, enableBar bool) *HlsDl {
	if filename == "" {
		filename = getFilename()
	}
	hlsdl := &HlsDl{
		hlsURL:    hlsURL,
		dir:       dir,
		client:    resty.New(),
		workers:   workers,
		enableBar: enableBar,
		headers:   headers,
		filename:  filename,
		startTime: time.Now().UnixMilli(),
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
	hlsDl.client.SetRetryCount(5).SetRetryWaitTime(time.Second)
	resp, err := hlsDl.client.R().SetHeaders(hlsDl.headers).SetOutput(segment.Path).Get(segment.URI)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.Status())
	}
	return nil
}

func (hlsDl *HlsDl) downloadSegments(segmentsDir string, segments []*Segment) error {
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
			segName := fmt.Sprintf("Seg%d.ts", segment.SeqId)
			segment.Path = filepath.Join(segmentsDir, segName)
			select {
			case segmentChan <- segment:
			case <-quitChan:
				return
			}
		}
	}()
	hlsDl.segTotal = int64(len(segments))
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
			} else {
				atomic.AddInt64(&hlsDl.segCurrent, 1)
			}
		}
	}

}

func (hlsDl *HlsDl) Download() (string, error) {
	segs, err := parseHlsSegments(hlsDl.hlsURL, hlsDl.headers)
	if err != nil {
		return "", err
	}
	segmentsDir := filepath.Join(hlsDl.dir, fmt.Sprintf("%d", hlsDl.startTime))
	if err := os.MkdirAll(segmentsDir, os.ModePerm); err != nil {
		return "", err
	}
	if err := hlsDl.downloadSegments(segmentsDir, segs); err != nil {
		return "", err
	}
	fp, err := hlsDl.join(segmentsDir, segs)
	if err != nil {
		return "", err
	}

	return fp, nil
}

func (hlsDl *HlsDl) getKey(segment *Segment) (key []byte, iv []byte, err error) {
	res, err := hlsDl.client.SetHeaders(hlsDl.headers).R().Get(segment.Key.URI)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode() != 200 {
		return nil, nil, errors.New("failed to get descryption key")
	}

	key = res.Body()
	iv = []byte(segment.Key.IV)
	if len(iv) == 0 {
		iv = defaultIV(segment.SeqId)
	}
	return
}

func (hlsDl *HlsDl) GetProgress() float64 {
	var current int64
	if hlsDl.enableBar {
		current = hlsDl.bar.Get()
	} else {
		current = atomic.LoadInt64(&hlsDl.segCurrent)
	}
	return float64(current) / float64(hlsDl.segTotal)
}

func (hlsDl *HlsDl) join(segmentsDir string, segments []*Segment) (string, error) {
	log.Println("Joining segments")

	outFile := filepath.Join(hlsDl.dir, hlsDl.filename)

	// Create the output file
	f, err := os.Create(outFile)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Sort the segments by SeqId to ensure correct order
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].SeqId < segments[j].SeqId
	})

	// Clean up the segmentsDir after processing
	defer func() {
		if err := os.RemoveAll(segmentsDir); err != nil {
			log.Printf("failed to remove segments directory: %v", err)
		}
	}()

	// Process each segment one by one
	for _, segment := range segments {
		if segment.Key != nil {
			key, iv, err := hlsDl.getKey(segment)
			if err != nil {
				return "", err
			}
			// Decrypt the segment and write it directly to the output file
			if err := decryptAES128Stream(segment.Path, f, key, iv); err != nil {
				return "", err
			}
		} else {
			if err := decodeStream(segment.Path, f); err != nil {
				return "", err
			}
		}

		// Remove the segment file after processing
		if err := os.Remove(segment.Path); err != nil {
			log.Printf("failed to remove segment file (%s): %v", segment.Path, err)
		}
	}

	return outFile, nil
}
