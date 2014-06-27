package layout

type Padded64Int64 struct {
	before [7]int64
	Value  int64
	after  [8]int64
}
