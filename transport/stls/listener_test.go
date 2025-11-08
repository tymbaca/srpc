package stls_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/transport/stls"
)

func TestNewListener(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		backing srpc.Listener
		key     *ecdh.PrivateKey
		wantErr bool
	}{
		{
			name: "good",
			key: func() *ecdh.PrivateKey {
				key, err := ecdh.P256().GenerateKey(rand.Reader)
				require.NoError(t, err)
				return key
			}(),
			wantErr: false,
		},
		{
			name: "bad 1",
			key: func() *ecdh.PrivateKey {
				key, err := ecdh.P384().GenerateKey(rand.Reader)
				require.NoError(t, err)
				return key
			}(),
			wantErr: true,
		},
		{
			name: "bad 2",
			key: func() *ecdh.PrivateKey {
				key, err := ecdh.P521().GenerateKey(rand.Reader)
				require.NoError(t, err)
				return key
			}(),
			wantErr: true,
		},
		{
			name: "bad 3",
			key: func() *ecdh.PrivateKey {
				key, err := ecdh.X25519().GenerateKey(rand.Reader)
				require.NoError(t, err)
				return key
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := stls.NewListener(tt.backing, tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewListener() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewListener() succeeded unexpectedly")
			}
			if got == nil {
				t.Errorf("NewListener() = %v, want normal value", got)
			}
		})
	}
}
