package request_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"
	"spitfiregg.httpFromScratch.httpieee/internal/request"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {

	t.Run("standard headers test", func(t *testing.T) {

		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}

		fmt.Println()
		fmt.Println("1st iteration")
		fmt.Println()
		r, err := request.RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.Equal(t, "localhost:42069", r.Header.Header["host"])

		require.Equal(t, "curl/7.81.0", r.Header.Header["user-agent"])
		require.Equal(t, "*/*", r.Header.Header["accept"])
	})

	t.Run("with content length", func(t *testing.T) {

		fmt.Println()
		fmt.Println("2nd iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +

				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
		}
		r, err := request.RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	t.Run("with partial content", func(t *testing.T) {
		fmt.Println()
		fmt.Println("3rd iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			numBytesPerRead: 3,
		}
		_, err := request.RequestFromReader(reader)
		require.Error(t, err)

	})

	t.Run("body shorter than reported content length", func(t *testing.T) {
		fmt.Println()
		fmt.Println("4th iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			numBytesPerRead: 3,
		}
		_, err := request.RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("standard Body", func(t *testing.T) {
		fmt.Println()
		fmt.Println("5th iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
		}
		r, err := request.RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	t.Run("empty Body, 0 reported content length", func(t *testing.T) {
		fmt.Println()
		fmt.Println("6th iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		}
		r, err := request.RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	t.Run("empty Body, no reported content length", func(t *testing.T) {
		fmt.Println()
		fmt.Println("7th iteration")
		fmt.Println()
		reader := &chunkReader{
			data: "GET /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		}
		r, err := request.RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

}
