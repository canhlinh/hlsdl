package hlsdl

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/grafov/m3u8"
)

type SegmentPuller struct {
	Segment *Segment
	Err     error
}

func pullSegment(hlsURL string, quitSignal chan os.Signal) chan *SegmentPuller {
	c := make(chan *SegmentPuller)

	go func() {
		defer close(c)

		baseURL, err := url.Parse(hlsURL)
		if err != nil {
			c <- &SegmentPuller{Err: err}
			return
		}

		pulledSegment := map[uint64]bool{}

		for {
			p, t, err := getM3u8ListType(hlsURL, nil)
			if err != nil {
				c <- &SegmentPuller{Err: err}
				return
			}
			if t != m3u8.MEDIA {
				c <- &SegmentPuller{Err: errors.New("No support the m3u8 format")}
				return
			}

			mediaList := p.(*m3u8.MediaPlaylist)
			if mediaList.Closed {
				c <- &SegmentPuller{Err: errors.New("The stream has been closed")}
			}

			duration := time.NewTicker(time.Duration(mediaList.TargetDuration) * time.Second)

			for _, seg := range mediaList.Segments {
				if seg == nil {
					continue
				}

				if pulledSegment[seg.SeqId] {
					continue
				} else {
					pulledSegment[seg.SeqId] = true
				}

				if !strings.Contains(seg.URI, "http") {
					segmentURL, err := baseURL.Parse(seg.URI)
					if err != nil {
						c <- &SegmentPuller{Err: err}
						return
					}

					seg.URI = segmentURL.String()
				}

				if seg.Key == nil && mediaList.Key != nil && mediaList.Key.Method != "NONE" {
					seg.Key = mediaList.Key
				}

				if seg.Key != nil {
					if seg.Key.Method == "NONE" {
						seg.Key = nil
					} else if !strings.Contains(seg.Key.URI, "http") {
						keyURL, err := baseURL.Parse(seg.Key.URI)
						if err != nil {
							c <- &SegmentPuller{Err: err}
							return
						}
						seg.Key.URI = keyURL.String()
					}
				}

				select {
				case c <- &SegmentPuller{Segment: &Segment{MediaSegment: seg}}:
				case <-quitSignal:
					log.Println("Got stop signal")
					return
				}
			}

			<-duration.C
		}
	}()

	return c
}
