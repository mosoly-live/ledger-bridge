package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnlyLettersValidator(t *testing.T) {
	testCases := []struct {
		name              string
		s                 string
		expectErr         bool
		expectedErrString string
	}{
		{
			name:      "English letters",
			s:         "John Doe",
			expectErr: false,
		},
		{
			name:              "emoji",
			s:                 "John♥",
			expectErr:         true,
			expectedErrString: " in body should match 'only letters'",
		},
		{
			name:              "symbols",
			s:                 "John$/=?",
			expectErr:         true,
			expectedErrString: " in body should match 'only letters'",
		},
		{
			name:              "trailing whitespace",
			s:                 "John ",
			expectErr:         true,
			expectedErrString: " in body should match 'only letters'",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)
			ol := OnlyLetters(testCase.s)
			err := ol.Validate(nil)
			if testCase.expectErr {
				r.Contains(err.Error(), testCase.expectedErrString)
			} else {
				r.NoError(err)
			}
		})
	}
}
