package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriceValidator(t *testing.T) {
	testCases := []struct {
		name      string
		number    float64
		expectErr bool
	}{
		{
			name:      "one cent",
			number:    0.01,
			expectErr: false,
		},
		{
			name:      "one euro one cent",
			number:    1.01,
			expectErr: false,
		},
		{
			name:      "too many decimals",
			number:    0.001,
			expectErr: true,
		},
		{
			name:      "zero with correct decimals",
			number:    0.00,
			expectErr: true,
		},
		{
			name:      "arbitrary number with correct decimals",
			number:    123.45,
			expectErr: false,
		},
		{
			name:      "arbitrary number with incorrect decimals",
			number:    123.45678,
			expectErr: true,
		},
		{
			name:      "negative number with correct decimals",
			number:    -123.45,
			expectErr: true,
		},
		{
			name:      "negative number with incorrect decimals",
			number:    -123.456,
			expectErr: true,
		},
		{
			name:      "positive number with too many decimals",
			number:    0.0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001,
			expectErr: true,
		},
		{
			name:      "negative number with too many decimals",
			number:    -0.0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001,
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)
			price := Price(testCase.number)
			err := price.Validate(nil)
			if testCase.expectErr {
				r.NotNil(err)
			} else {
				r.NoError(err)
			}
		})
	}
}
