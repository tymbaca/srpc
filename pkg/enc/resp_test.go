package enc

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRep(t *testing.T) {
	tests := []struct {
		name  string
		input Response
	}{
		{
			name: "ok with body",
			input: Response{
				StatusCode: StatusOK,
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Error: nil,
				Body:  bytes.NewBufferString("this is the input"),
			},
		},
		{
			name: "ok without body",
			input: Response{
				StatusCode: StatusOK,
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Error: nil,
				Body:  nil,
			},
		},
		{
			name: "ok empty buffer body",
			input: Response{
				StatusCode: StatusOK,
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Error: nil,
				Body:  bytes.NewBufferString(""),
			},
		},
		{
			name: "error",
			input: Response{
				StatusCode: StatusInternalError,
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Error: errors.New("some error"),
				Body:  nil,
			},
		},
		{
			name: "error with empty text",
			input: Response{
				StatusCode: StatusInternalError,
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Error: errors.New(""),
				Body:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputBody, outputBody []byte

			if tt.input.Body != nil {
				inputBody = tt.input.Body.(*bytes.Buffer).Bytes()
			}

			buf := new(bytes.Buffer)
			err := WriteResponse(buf, tt.input)
			require.NoError(t, err)

			output, err := ReadResponse(buf)
			require.NoError(t, err)

			if tt.input.Body != nil {
				outputBody, err = io.ReadAll(output.Body)
				require.NoError(t, err)
			}

			tt.input.Body = nil
			output.Body = nil
			require.Equal(t, tt.input, output)
			require.Equal(t, string(inputBody), string(outputBody))
		})
	}
}
