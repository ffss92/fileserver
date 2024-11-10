package fileserver

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestCalculateETag(t *testing.T) {
	tests := []struct {
		name         string
		input        io.Reader
		expectedETag string
		fails        bool
	}{
		{
			name:         "with value",
			input:        strings.NewReader("Hello, world!"),
			expectedETag: `"6cd3556deb0da54bca060b4c39479839"`, // pre-calculated hash
			fails:        false,
		},
		{
			name:         "without value",
			input:        strings.NewReader(""),
			expectedETag: `"d41d8cd98f00b204e9800998ecf8427e"`, // empty string hash
			fails:        false,
		},
		{
			name:         "error reader",
			input:        &errorReader{},
			expectedETag: "",
			fails:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etag, err := calculateETag(tt.input)
			if (err != nil) != tt.fails {
				t.Fatalf("expected error to be %v but got %v", tt.fails, err)
			}
			if etag != tt.expectedETag {
				t.Errorf("expected etag to be %v but got %v", tt.expectedETag, etag)
			}
		})
	}
}
