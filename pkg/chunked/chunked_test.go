package chunked

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotBuffered(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var wg sync.WaitGroup

		client, server := net.Pipe()

		clientChunked := NewWriter(client)
		serverChunked := NewReader(server)

		data := [][]byte{
			[]byte("yo"),
			[]byte("helloworld"),
			{},
		}

		data[2] = make([]byte, math.MaxUint16*3+600)
		rand.Read(data[2])

		wg.Add(1)
		go func() {
			for _, c := range data {
				n, err := clientChunked.Write(c)
				require.Equal(t, len(c), n)
				require.NoError(t, err)
			}
			err := clientChunked.Close()
			require.NoError(t, err)
		}()

		var got []byte
		buf := make([]byte, 512)

		n, err := serverChunked.Read(buf)
		require.Equal(t, len(data[0]), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		n, err = serverChunked.Read(buf)
		require.Equal(t, len(data[1]), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		// third chunk is big, so at least his first will fill buf entirely
		n, err = serverChunked.Read(buf)
		require.Equal(t, len(buf), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		for {
			n, err := serverChunked.Read(buf)
			got = append(got, buf[:n]...)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				require.FailNow(t, "reader got non-EOF error: %s", err)
			}
		}

		expected := bytes.Join(data, nil)
		require.Equal(t, expected, got)
	})
}
