package hlsdl

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

type Recorder struct {
	client *http.Client
	dir    string
	url    string
	runing bool
}

func NewRecorder(url string, dir string) *Recorder {
	return &Recorder{
		url:    url,
		dir:    dir,
		client: &http.Client{},
	}
}

func (r *Recorder) Start() (string, error) {
	log.Println("Start record live streaming movie...")

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt)
	puller := pullSegment(r.url, quitSignal)

	filePath := filepath.Join(r.dir, "video.ts")
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for segment := range puller {
		if segment.Err != nil {
			return "", segment.Err
		}

		segmentData, err := r.getSegmentData(segment.Segment)
		if err != nil {
			return "", segment.Err
		}

		if _, err := file.Write(segmentData); err != nil {
			return "", segment.Err
		}

		log.Println("Recorded segment ", segment.Segment.SeqId)
	}

	return filePath, nil
}

func (r *Recorder) getSegmentData(segment *Segment) ([]byte, error) {

	res, err := r.client.Get(segment.URI)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if segment.Key != nil {
		key, iv, err := r.getKey(segment)
		if err != nil {
			return nil, err
		}
		data, err = AES128Decrypt(data, key, iv)
	}

	syncByte := uint8(71) //0x47
	bLen := len(data)
	for j := 0; j < bLen; j++ {
		if data[j] == syncByte {
			data = data[j:]
			break
		}
	}

	return data, nil
}

func (r *Recorder) getKey(segment *Segment) (key []byte, iv []byte, err error) {
	res, err := r.client.Get(segment.Key.URI)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != 200 {
		return nil, nil, errors.New("Failed to get descryption key")
	}

	key, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	iv = []byte(segment.Key.IV)
	return
}
