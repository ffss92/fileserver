package fileserver

import (
	"net/http"
	"testing"
)

func TestAcceptsGzip(t *testing.T) {
	testCases := []struct {
		name       string
		newRequest func() (*http.Request, error)
		expected   bool
	}{
		{
			name: "accepts",
			newRequest: func() (*http.Request, error) {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Accept-Encoding", "br, gzip")
				return req, nil
			},
			expected: true,
		},
		{
			name: "dont accept",
			newRequest: func() (*http.Request, error) {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Accept-Encoding", "br")
				return req, nil
			},
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.newRequest()
			if err != nil {
				t.Fatalf("unexpected error creating request: %s", err)
			}
			result := acceptsGzip(req)
			if result != tt.expected {
				t.Errorf("expected result to be %t but got %t", tt.expected, result)
			}
		})
	}
}
