package datastore

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

const (
	StringType = 0
	Int64Type  = 1
)

type Entry struct {
	key       string
	valueType int
	value     interface{}
}

func NewEntry(key string, valueType int, value interface{}) *Entry {
	return &Entry{key: key, valueType: valueType, value: value}
}

func getLength(key string, valueType int, value interface{}) int64 {
	kl := len(key)
	vl := 0
	if valueType == StringType {
		vl = len(value.(string))
	} else if valueType == Int64Type {
		vl = 8 // size of int64
	}
	return int64(kl + vl + 16) // 16 bytes for size (4), type (4), key length (4), value length (4)
}

func (e *Entry) Encode() []byte {
	kl := len(e.key)
	vl := 0
	if e.valueType == StringType {
		vl = len(e.value.(string))
	} else if e.valueType == Int64Type {
		vl = 8 // size of int64
	}
	size := kl + vl + 16
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(size))
	binary.LittleEndian.PutUint32(res[4:], uint32(e.valueType))
	binary.LittleEndian.PutUint32(res[8:], uint32(kl))
	copy(res[12:], e.key)
	if e.valueType == StringType {
		binary.LittleEndian.PutUint32(res[kl+12:], uint32(vl))
		copy(res[kl+16:], e.value.(string))
	} else if e.valueType == Int64Type {
		binary.LittleEndian.PutUint32(res[kl+12:], 8)
		binary.LittleEndian.PutUint64(res[kl+16:], uint64(e.value.(int64)))
	}
	return res
}

func (e *Entry) GetLength() int64 {
	return getLength(e.key, e.valueType, e.value)
}

func (e *Entry) Decode(input []byte) error {
	size := binary.LittleEndian.Uint32(input)
	if len(input) != int(size) {
		return fmt.Errorf("input size does not match encoded size")
	}
	e.valueType = int(binary.LittleEndian.Uint32(input[4:]))
	kl := binary.LittleEndian.Uint32(input[8:])
	keyBuf := make([]byte, kl)
	copy(keyBuf, input[12:kl+12])
	e.key = string(keyBuf)
	vl := binary.LittleEndian.Uint32(input[kl+12:])
	if e.valueType == StringType {
		valBuf := make([]byte, vl)
		copy(valBuf, input[kl+16:kl+16+vl])
		e.value = string(valBuf)
	} else if e.valueType == Int64Type {
		if vl != 8 {
			return fmt.Errorf("invalid int64 value length")
		}
		e.value = int64(binary.LittleEndian.Uint64(input[kl+16:kl+24]))
	} else {
		return fmt.Errorf("unknown value type")
	}
	return nil
}

func readValue(in *bufio.Reader) (int, interface{}, error) {
	header, err := in.Peek(12)
	if err != nil {
		return 0, nil, err
	}
	valueType := int(binary.LittleEndian.Uint32(header[4:]))
	keySize := int(binary.LittleEndian.Uint32(header[8:]))
	_, err = in.Discard(keySize + 12)
	if err != nil {
		return 0, nil, err
	}

	header, err = in.Peek(4)
	if err != nil {
		return 0, nil, err
	}
	valSize := int(binary.LittleEndian.Uint32(header))
	_, err = in.Discard(4)
	if err != nil {
		return 0, nil, err
	}

	if valueType == StringType {
		data := make([]byte, valSize)
		n, err := in.Read(data)
		if err != nil {
			return 0, nil, err
		}
		if n != valSize {
			return 0, nil, fmt.Errorf("can't read value bytes (read %d, expected %d)", n, valSize)
		}
		return valueType, string(data), nil
	} else if valueType == Int64Type {
		if valSize != 8 {
			return 0, nil, fmt.Errorf("invalid int64 value length")
		}
		data := make([]byte, 8)
		n, err := in.Read(data)
		if err != nil {
			return 0, nil, err
		}
		if n != 8 {
			return 0, nil, fmt.Errorf("can't read int64 bytes (read %d, expected %d)", n, 8)
		}
		return valueType, int64(binary.LittleEndian.Uint64(data)), nil
	}
	return 0, nil, fmt.Errorf("unknown value type")
}
