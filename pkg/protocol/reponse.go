package protocol

const TypeResponse = 'r'

type Response struct {
	Code    int
	Message string
	Data    interface{}
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
func SuccessWithData(data interface{}) *Response {
	result := Success()
	result.Data = data
	return result
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
