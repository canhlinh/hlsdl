package hlsdl

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafov/m3u8"
)

func parseHlsSegments(hlsURL string) ([]*Segment, error) {
	baseURL, err := url.Parse(hlsURL)
	if err != nil {
		return nil, err
	}

	res, err := http.Get(hlsURL)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	p, t, err := m3u8.DecodeFrom(res.Body, true)
	if err = p.DecodeFrom(res.Body, false); err != nil {
		return nil, err
	}

	if t != m3u8.MEDIA {
		return nil, errors.New("No support the m3u8 format")
	}

	mediaList := p.(*m3u8.MediaPlaylist)
	segments := []*Segment{}
	for _, seg := range mediaList.Segments {
		if seg == nil {
			continue
		}

		if !strings.Contains(seg.URI, "http") {
			segmentURL, err := baseURL.Parse(seg.URI)
			if err != nil {
				return nil, err
			}

			seg.URI = segmentURL.String()

			segment := &Segment{MediaSegment: seg}
			segments = append(segments, segment)
		}
	}

	return segments, nil
}
