package errwrap

import "fmt"

func PanicToError(f func()) (e error) {
	defer func() {
		if err := recover(); err != nil {
			e = fmt.Errorf("panic error:%v", err)
		}
	}()
	f()
	return
}
