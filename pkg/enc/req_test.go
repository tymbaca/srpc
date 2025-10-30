package enc

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReq(t *testing.T) {
	tests := []struct {
		name    string
		input   Request
		encoder *Encoder
	}{
		{
			name: "full",
			input: Request{
				Version:       Version{1, 0, 0},
				ServiceMethod: ServiceMethod(NewString("testService.testMethod")),
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Body: bytes.NewBufferString("this is the input"),
			},
		},
		{
			name: "no body",
			input: Request{
				Version:       Version{1, 0, 0},
				ServiceMethod: ServiceMethod(NewString("testService.testMethod")),
				Metadata: NewMetadata(map[string][]string{
					"k1": {"v1", "v2"},
					"k2": {"v3", "v4"},
					"k3": {},
				}),
				Body: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputBody, outputBody []byte

			if tt.input.Body != nil {
				inputBody = tt.input.Body.(*bytes.Buffer).Bytes()
			}

			var e *Encoder
			if tt.encoder != nil {
				e = tt.encoder
			} else {
				e = &Encoder{Version: tt.input.Version, IgnoreVersion: false}
			}

			buf := new(bytes.Buffer)
			err := e.WriteRequest(buf, tt.input)
			require.NoError(t, err)

			output, err := e.ReadRequest(buf)
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
