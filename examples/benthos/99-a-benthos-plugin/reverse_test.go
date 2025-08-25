// See https://github.com/benthosdev/benthos-plugin-example/blob/master/processor/reverse_test.go
package processor

import (
	"context"
	"testing"

	"github.com/warpstreamlabs/bento/public/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReverseProcessor(t *testing.T) {
	testCases := []struct {
		desc       string
		msg        string
		numResults int
		want       string
	}{
		{"simple string", "hello world", 1, "dlrow olleh"},
		{"palindrome", "wooooow", 1, "wooooow"},
	}

	// It's safe to pass nil in place of a logger for testing purposes
	revProc := newReverseProcessor(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := revProc.Process(context.Background(), service.NewMessage([]byte(tc.msg)))
			require.NoError(t, err)
			require.Len(t, result, tc.numResults)

			resBytes, err := result[0].AsBytes()
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(resBytes))
		})
	}
}
