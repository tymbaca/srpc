package chunked

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func xrun(t *testing.T, name string,
	prep func(r io.Reader, w io.Writer) (io.Reader, io.WriteCloser),
	input [][]byte,
	check func(t *testing.T, input [][]byte, r io.Reader, writerExited *atomic.Bool),
) {
	// do nothing
}

func run(t *testing.T, name string,
	prep func(r io.Reader, w io.Writer) (io.Reader, io.WriteCloser),
	input [][]byte,
	check func(t *testing.T, input [][]byte, r io.Reader, writerExited *atomic.Bool),
) {
	t.Run(name, func(t *testing.T) {
		var wg sync.WaitGroup

		client, server := net.Pipe()
		r, w := prep(server, client)

		var writerExited atomic.Bool

		wg.Add(1)
		go func() {
			defer func() {
				writerExited.Store(true)
			}()

			for _, c := range input {
				n, err := w.Write(c)
				require.Equal(t, len(c), n)
				require.NoError(t, err)
			}
			err := w.Close()
			require.NoError(t, err)
		}()

		check(t, input, r, &writerExited) // check must drain the reader

		time.Sleep(5 * time.Millisecond) // to be sure
		require.True(t, writerExited.Load(), "writer didn't exit")
	})
}

func TestChunked(t *testing.T) {
	noBufPrep := func(r io.Reader, w io.Writer) (io.Reader, io.WriteCloser) {
		return NewReader(r), NewWriter(w)
	}

	testData1 := [][]byte{
		[]byte("yo"),
		[]byte("helloworld"),
		{},
	}
	testData1[2] = make([]byte, math.MaxUint16*3+600)
	rand.Read(testData1[2])

	run(t, "not buffered | debug", noBufPrep, [][]byte{[]byte("yo")}, func(t *testing.T, input [][]byte, r io.Reader, _ *atomic.Bool) {
		var got []byte
		buf := make([]byte, 512)

		n, err := r.Read(buf)
		require.Equal(t, 2, n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		n, err = r.Read(buf)
		require.Equal(t, 0, n)
		require.ErrorIs(t, err, io.EOF)

		expected := bytes.Join(input, nil)
		require.Equal(t, expected, got)
	})

	run(t, "not buffered", noBufPrep, testData1, func(t *testing.T, input [][]byte, r io.Reader, _ *atomic.Bool) {
		var got []byte
		buf := make([]byte, 512)

		n, err := r.Read(buf)
		require.Equal(t, len(testData1[0]), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		n, err = r.Read(buf)
		require.Equal(t, len(testData1[1]), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		// third chunk is big, so at least his first will fill buf entirely
		n, err = r.Read(buf)
		require.Equal(t, len(buf), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		for {
			n, err := r.Read(buf)
			got = append(got, buf[:n]...)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				require.FailNow(t, "reader got non-EOF error: %s", err)
			}
		}

		expected := bytes.Join(input, nil)
		require.Equal(t, expected, got)
	})

	run(t, "not buffered | empty", noBufPrep, nil, func(t *testing.T, input [][]byte, r io.Reader, _ *atomic.Bool) {
		buf := make([]byte, 512)

		n, err := r.Read(buf)
		require.Equal(t, 0, n)
		require.ErrorIs(t, err, io.EOF)
	})

	bufPrep := func(r io.Reader, w io.Writer) (io.Reader, io.WriteCloser) {
		return NewReader(r), NewBufferWriter(w)
	}

	run(t, "buffered", bufPrep, testData1, func(t *testing.T, input [][]byte, r io.Reader, _ *atomic.Bool) {
		var got []byte
		buf := make([]byte, 512)

		n, err := r.Read(buf)
		require.Equal(t, len(buf), n)
		require.NoError(t, err)
		got = append(got, buf[:n]...)

		for {
			n, err := r.Read(buf)
			got = append(got, buf[:n]...)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				require.FailNow(t, "reader got non-EOF error: %s", err)
			}
		}

		expected := bytes.Join(input, nil)
		require.Equal(t, expected, got)
	})

	run(t, "buffered | json", bufPrep, sli([]byte(`{"Foo": "Bar"}`)), func(t *testing.T, input [][]byte, r io.Reader, _ *atomic.Bool) {
		type FooBar struct {
			Foo string
		}

		want := FooBar{Foo: "Bar"}

		var got FooBar
		err := json.NewDecoder(r).Decode(&got)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func sli[T any](vs ...T) []T {
	return vs
}
