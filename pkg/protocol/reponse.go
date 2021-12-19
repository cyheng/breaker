package protocol

const TypeResponse = 'r'

type Response struct {
	Code    int
	Message string
}

func (n *Response) Type() byte {
	return TypeResponse
}
func Success() *Response {
	return &Response{
		Code:    0,
		Message: "success",
	}
}
func Error(msg string) *Response {
	return &Response{
		Code:    -1,
		Message: msg,
	}
}
func init() {
	RegisterCommand(&Response{})
}
