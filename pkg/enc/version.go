package enc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
)

type Version struct {
	Major, Minor, Patch uint16
}

func (v Version) CompatibleWith(other Version) bool {
	return v.Major == other.Major
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (e *Encoder) checkVersion(r io.Reader) (Version, error) {
	other, err := readVersion(r)
	if err != nil {
		return Version{}, err
	}

	if !e.IgnoreVersion && !e.Version.CompatibleWith(other) {
		return other, incompatibleVersionError(e.Version, other)
	}

	return other, nil
}

func readVersion(r io.Reader) (Version, error) {
	var ver Version
	if err := sbinary.NewDecoder(r).Decode(&ver, binary.BigEndian); err != nil {
		return Version{}, fmt.Errorf("decode version: %w", err)
	}

	return ver, nil
}

func writeVersion(w io.Writer, ver Version) error {
	if err := sbinary.NewEncoder(w).Encode(ver, binary.BigEndian); err != nil {
		return fmt.Errorf("encode version: %w", err)
	}

	return nil
}

var ErrIncompatibleVersion = errors.New("version is not compatible")

func incompatibleVersionError(my Version, other Version) error {
	return fmt.Errorf("my version is %s, got %s: %w", my, other, ErrIncompatibleVersion)
}
