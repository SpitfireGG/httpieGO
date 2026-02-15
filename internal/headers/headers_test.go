package headers_test

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/require"
	. "spitfiregg.httpFromScratch.httpieee/internal/headers"
)

func Test_Parse(t *testing.T) {
	headers := NewHeaders()
	data01 := []byte("Set-Person: lane-loves-go\r\n")
	c1, done, err := headers.Parse(data01)
	require.NoError(t, err)
	require.Equal(t, headers.Header["set-person"], "lane-loves-go", "must be the key for the header field")
	require.Equal(t, c1, 27)
	assert.False(t, done)

	data02 := []byte("Set-Person: prime loves zig\r\n")
	c2, done2, err2 := headers.Parse(data02)
	require.NoError(t, err2)

	t.Run("should return true", func(t *testing.T) {
		require.Equal(t, headers.Header["set-person"], "lane-loves-go,prime loves zig")
		require.Equal(t, c2, 29)
	})
	assert.False(t, done2)

	data03 := []byte("Set-Person: tj-loves-ocaml\r\n")
	c3, done3, err3 := headers.Parse(data03)
	require.NoError(t, err3)

	t.Run("should return true", func(t *testing.T) {
		require.Equal(t, headers.Header["set-person"], "lane-loves-go,prime loves zig,tj-loves-ocaml")
		require.Equal(t, c3, 28)
	})
	assert.False(t, done3)

}
