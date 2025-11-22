package status_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc/status"
)

func TestStatusFromError(t *testing.T) {
	_, ok := status.FromError(nil)
	require.False(t, ok)

	_, ok = status.FromError(errors.New("some error"))
	require.False(t, ok)

	code, ok := status.FromError(fmt.Errorf("wrapped err: %w", status.Error(status.OK, "test")))
	require.True(t, ok)
	require.Equal(t, status.OK, code)

	code, ok = status.FromError(fmt.Errorf("wrapped err: %w", status.Error(status.InvalidArgument, "test")))
	require.True(t, ok)
	require.Equal(t, status.InvalidArgument, code)

	code, ok = status.FromError(fmt.Errorf("wrapped err: %w", status.Errorf(status.InvalidArgument, "test some int: %d", 777)))
	require.True(t, ok)
	require.Equal(t, status.InvalidArgument, code)
}
