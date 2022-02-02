package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

var (
	ErrMsgType      = errors.New("message type error")
	ErrMaxMsgLength = errors.New("message length exceed the limit")
	ErrMsgLength    = errors.New("message length error")
	ErrMsgFormat    = errors.New("message format error")
)

var (
	defaultMaxMsgLength int64   = 10240
	MsgManager          *MsgCtl = NewMsgCtl()
)

type MsgCtl struct {
	typeMap     map[byte]reflect.Type
	typeByteMap map[reflect.Type]byte

	maxMsgLength int64
}

func (msgCtl *MsgCtl) SetMaxMsgLength(length int64) {
	msgCtl.maxMsgLength = length
}
func NewMsgCtl() *MsgCtl {
	return &MsgCtl{
		typeMap:      make(map[byte]reflect.Type),
		typeByteMap:  make(map[reflect.Type]byte),
		maxMsgLength: defaultMaxMsgLength,
	}
}

func RegisterCommand(msg Command) {
	_, ok := MsgManager.typeMap[msg.Type()]
	if ok {
		panic("message type:" + string(msg.Type()) + " has been register")
	}
	MsgManager.typeMap[msg.Type()] = reflect.TypeOf(msg)
	MsgManager.typeByteMap[reflect.TypeOf(msg)] = msg.Type()
}

func readMsg(c io.Reader) (typeByte byte, buffer []byte, err error) {
	buffer = make([]byte, 1)

	_, err = c.Read(buffer)
	if err != nil {
		return
	}
	typeByte = buffer[0]
	if _, ok := MsgManager.typeMap[typeByte]; !ok {
		err = ErrMsgType
		return
	}

	var length int64
	err = binary.Read(c, binary.BigEndian, &length)
	if err != nil {
		return
	}
	if length > MsgManager.maxMsgLength {
		err = ErrMaxMsgLength
		return
	} else if length < 0 {
		err = ErrMsgLength
		return
	}

	buffer = make([]byte, length)
	n, err := io.ReadFull(c, buffer)
	if err != nil {
		return
	}

	if int64(n) != length {
		err = ErrMsgFormat
	}
	return
}

func ReadMsg(c io.Reader) (msg Command, err error) {
	typeByte, buffer, err := readMsg(c)
	if err != nil {
		return
	}
	t, ok := MsgManager.typeMap[typeByte]
	if !ok {
		err = ErrMsgType
		return
	}
	//指针类型获取真正type需要调用Elem
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	msg = reflect.New(t).Interface().(Command)
	err = json.Unmarshal(buffer, &msg)
	return
}
func CmdToBytes(msg Command) ([]byte, error) {
	typeByte := msg.Type()

	content, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	buffer.WriteByte(typeByte)
	_ = binary.Write(buffer, binary.BigEndian, int64(len(content)))
	buffer.Write(content)
	return buffer.Bytes(), nil
}
func WriteMsg(c io.Writer, msg Command) (err error) {
	buf, err := CmdToBytes(msg)
	if err != nil {
		return err
	}
	if _, err = c.Write(buf); err != nil {
		return
	}
	return nil
}
