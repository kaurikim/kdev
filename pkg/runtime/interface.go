package runtime

type Object interface {
	Encode() (string, error)
	Decode([]byte) error
	Clone() Object
}
