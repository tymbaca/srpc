package srpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tymbaca/srpc/pkg/enc"
)

func TestStatusFromError(t *testing.T) {
	tests := []struct {
		err  error
		want Status
	}{
		{nil, enc.StatusOK},
		{errors.New("some error"), enc.StatusInternalError},
		{fmt.Errorf("%w: something else: %w", statusError{status: enc.StatusInternalError}, errors.New("some error")), enc.StatusInternalError},
		{fmt.Errorf("%w: something else: %w", statusError{status: enc.StatusOK}, errors.New("some error")), enc.StatusOK},
		{fmt.Errorf("%w: something else: %w", statusError{status: enc.StatusInvalidArgument}, errors.New("some error")), enc.StatusInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.err), func(t *testing.T) {
			got := StatusFromError(tt.err)
			if got != tt.want {
				t.Errorf("StatusFromError() = %v, want %v", got, tt.want)
			}
		})
	}
}
