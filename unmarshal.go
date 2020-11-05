package gopymarshal

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const (
	CODE_NONE      = 'N' //None
	CODE_INT       = 'i' //integer
	CODE_INT2      = 'c' //integer2
	CODE_FLOAT     = 'g' //float
	CODE_STRING    = 's' //string
	CODE_UNICODE   = 'u' //unicode string
	CODE_TSTRING   = 't' //tstring?
	CODE_TUPLE     = '(' //tuple
	CODE_LIST      = '[' //list
	CODE_DICT      = '{' //dict
	CODE_STOP      = '0'
	DICT_INIT_SIZE = 64
)

var (
	ERR_PARSE        = errors.New("invalid data")
	ERR_UNKNOWN_CODE = errors.New("unknown code")
)

// Unmarshal data serialized by python
func Unmarshal(r io.Reader) (ret interface{}, retErr error) {
	ret, retErr = Unmarshal2(r)
	return
}

// Unmarshal2 data serialized by python, returning the unused portion.
func Unmarshal2(input io.Reader) (ret interface{}, retErr error) {
	code := make([]byte, 1)
	_, err := input.Read(code)
	if nil != err {
		retErr = err
	}

	ret, retErr = unmarshal(code[0], input)
	return
}

func unmarshal(code byte, r io.Reader) (ret interface{}, retErr error) {
	switch code {
	case CODE_NONE:
		ret = nil
	case CODE_INT:
		fallthrough
	case CODE_INT2:
		ret, retErr = readInt32(r)
	case CODE_FLOAT:
		ret, retErr = readFloat64(r)
	case CODE_STRING:
		fallthrough
	case CODE_UNICODE:
		fallthrough
	case CODE_TSTRING:
		ret, retErr = readString(r)
	case CODE_TUPLE:
		fallthrough
	case CODE_LIST:
		ret, retErr = readList(r)
	case CODE_DICT:
		ret, retErr = readDict(r)
	default:
		retErr = ERR_UNKNOWN_CODE
	}

	return
}

func readInt32(buffer io.Reader) (ret int32, retErr error) {
	var tmp int32
	retErr = ERR_PARSE
	if retErr = binary.Read(buffer, binary.LittleEndian, &tmp); nil == retErr {
		ret = tmp
	}

	return
}

func readFloat64(input io.Reader) (ret float64, retErr error) {
	retErr = ERR_PARSE
	tmp := make([]byte, 8)
	if num, err := input.Read(tmp); nil == err && 8 == num {
		bits := binary.LittleEndian.Uint64(tmp)
		ret = math.Float64frombits(bits)
		retErr = nil
	}

	return
}

func readString(input io.Reader) (ret string, retErr error) {
	var strLen int32
	strLen = 0
	retErr = ERR_PARSE
	if err := binary.Read(input, binary.LittleEndian, &strLen); nil != err {
		retErr = err
		return
	}

	retErr = nil
	buf := make([]byte, strLen)
	input.Read(buf)
	ret = string(buf)
	return
}

func readList(input io.Reader) (ret []interface{}, retErr error) {
	var listSize int32
	if retErr = binary.Read(input, binary.LittleEndian, &listSize); nil != retErr {
		return
	}

	code := []byte{0}
	var err error
	var val interface{}
	ret = make([]interface{}, int(listSize))
	for idx := 0; idx < int(listSize); idx++ {
		_, err = input.Read(code)
		if nil != err {
			break
		}

		val, err = unmarshal(code[0], input)
		if nil != err {
			retErr = err
			break
		}
		ret = append(ret, val)
	} //end of read loop

	return
}

func readDict(input io.Reader) (ret map[interface{}]interface{}, retErr error) {
	code := []byte{0}
	var err error
	var key interface{}
	var val interface{}
	ret = make(map[interface{}]interface{})
	for {
		_, err = input.Read(code)
		if nil != err {
			break
		}

		if CODE_STOP == code[0] {
			break
		}

		key, err = unmarshal(code[0], input)
		if nil != err {
			retErr = err
			break
		}

		_, err = input.Read(code)
		if nil != err {
			break
		}

		val, err = unmarshal(code[0], input)
		if nil != err {
			retErr = err
			break
		}
		ret[key] = val
	} //end of read loop

	return
}
