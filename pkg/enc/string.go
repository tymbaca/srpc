package enc

type String struct {
	Len  uint32 `sbin:"lenof:Data"`
	Data string
}

func NewString(v string) String {
	return String{uint32(len(v)), v}
}
