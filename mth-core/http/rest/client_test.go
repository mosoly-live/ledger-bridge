package rest

import (
	"fmt"
	"testing"
)

type SampleData struct {
	endpoint Endpoint
	uri      string
	expected string
}

func TestEndpointPath(t *testing.T) {
	expected := "https://api.staging.monetha.io/tyk/keys/wrong_key?hello=world#abcdef"

	testCases := []SampleData{
		{
			endpoint: Endpoint{
				// no slash at the end
				rawURL: "https://api.staging.monetha.io",
			},
			// no slash in the beginning
			uri:      "tyk/keys/wrong_key?hello=world#abcdef",
			expected: expected,
		},

		{
			endpoint: Endpoint{
				// slash at the end
				rawURL: "https://api.staging.monetha.io/",
			},
			// slash in the beginning
			uri:      "/tyk/keys/wrong_key?hello=world#abcdef",
			expected: expected,
		},

		{
			endpoint: Endpoint{
				// two slashes at the end
				rawURL: "https://api.staging.monetha.io//",
			},
			// two slashes at the end
			uri:      "/tyk/keys/wrong_key//?hello=world#abcdef",
			expected: expected,
		},

		{
			endpoint: Endpoint{
				// username / password
				rawURL: "http://username:password@example.com/path//",
			},
			uri:      "/to//?one=1#abcdef",
			expected: "http://username:password@example.com/path/to?one=1#abcdef",
		},

		{
			endpoint: Endpoint{
				// invalid value will trigger error
				rawURL: "😂://example.com",
			},
			uri: "/hello/world",

			// stays unmodified
			expected: "😂://example.com",
		},
	}

	for _, sample := range testCases {
		sample.endpoint.Path(sample.uri)
		if sample.expected != sample.endpoint.rawURL {
			t.Error(fmt.Sprintf("Expected: %s, got: %s", sample.expected, sample.endpoint.rawURL))
		}
	}
}
