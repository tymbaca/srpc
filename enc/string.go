package enc

// String holds string.
type String struct {
	Len  uint32 `sbin:"lenof:Data"`
	Data string
}

// NewString creates [String] from string.
func NewString(v string) String {
	return String{uint32(len(v)), v}
}

func (s String) String() string {
	return s.Data
}
