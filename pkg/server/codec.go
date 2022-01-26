package server

// Codec is a generic codec for encoding and decoding data.
type Codec interface {
	// Encode encodes data into []byte.
	// Returns error when error occurred.
	Encode(v interface{}) ([]byte, error)

	// Decode decodes data into v.
	// Returns error when error occurred.
	Decode(data []byte, v interface{}) error
}

type DefaultCodec struct{}

func NewDefaultCodec() *DefaultCodec {
	return &DefaultCodec{}
}

func (c * DefaultCodec) Encode(v interface{}) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (c *DefaultCodec) Decode(data []byte, v interface{}) error {
	//TODO implement me
	panic("implement me")
}

