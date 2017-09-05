// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package iendecode

import (
	"bytes"
	"encoding"
	"fmt"
	"regexp"
	"time"
)

func SingleToByte(thetype string, data interface{}) (b []byte, err error) {
	switch thetype {
	case "bool":
		b = BoolToBytes(data.(bool))
	case "byte":
		b = make([]byte, 1)
		b[0] = data.(byte)
	case "int8":
		b = Int8ToBytes(data.(int8))
	case "int16":
		b = Int16ToBytes(data.(int16))
	case "int32":
		b = Int32ToBytes(data.(int32))
	case "int64":
		b = Int64ToBytes(data.(int64))
	case "uint8":
		b = Uint8ToBytes(data.(uint8))
	case "uint16":
		b = Uint16ToBytes(data.(uint16))
	case "uint32":
		b = Uint32ToBytes(data.(uint32))
	case "uint64":
		b = Uint64ToBytes(data.(uint64))
	case "int":
		b = Int64ToBytes(data.(int64))
	case "uint":
		b = Uint64ToBytes(data.(uint64))
	case "float32":
		b, err = ToBinary(data)
		if err != nil {
			return
		}
	case "float64":
		b, err = ToBinary(data)
		if err != nil {
			return
		}
	case "complex64":
		b, err = ToBinary(data)
		if err != nil {
			return
		}
	case "complex128":
		b, err = ToBinary(data)
		if err != nil {
			return
		}
	case "string":
		b = []byte(data.(string))
	case "[]byte":
		b = data.([]byte)
	default:
		thedata, ok := data.(encoding.BinaryMarshaler)
		if ok == true {
			b, err = thedata.MarshalBinary()
			return
		} else {
			b, err = StructGobBytes(data)
			return
		}
	}
	return
}

func SliceToByte(thetype string, slice interface{}) (b []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	if t, _ := regexp.MatchString(`^\[\](bool|int|uint|int8|uint8|int16|uint16|int32|uint32|int64|uint64|float32|float64|complex46|complex128)`, thetype); t == true {
		b, err = ToBinary(slice)
		return
	} else {
		var b_buf bytes.Buffer

		if thetype == "[]string" {
			slicet, ok := slice.([]string)
			if ok == false {
				err = fmt.Errorf("Type Error.")
				return
			}
			thecount := int64(len(slicet))
			b_buf.Write(Int64ToBytes(thecount))
			for i := range slicet {
				str_b := []byte(slicet[i])
				str_b_len := int64(len(str_b))
				b_buf.Write(Int64ToBytes(str_b_len))
				b_buf.Write(str_b)
			}
		} else if thetype == "[]time.Time" {
			slicet, ok := slice.([]time.Time)
			if ok == false {
				err = fmt.Errorf("Type Error.")
				return
			}
			thecount := int64(len(slicet))
			b_buf.Write(Int64ToBytes(thecount))
			for i := range slicet {
				var d_b []byte
				d_b, err = slicet[i].MarshalBinary()
				if err != nil {
					return
				}
				d_b_len := int64(len(d_b))
				b_buf.Write(Int64ToBytes(d_b_len))
				b_buf.Write(d_b)
			}
		} else {
			var d_b []byte
			d_b, err = StructGobBytes(slice)
			if err != nil {
				return
			}
			d_b_len := int64(len(d_b))
			b_buf.Write(Int64ToBytes(d_b_len))
			b_buf.Write(d_b)
		}
		b = b_buf.Bytes()
		return
	}

}

func MapToByte(thetype string, themaps interface{}) (b []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	var b_buf bytes.Buffer
	switch thetype {
	case "map[string]string":
		themap := themaps.(map[string]string)
		thecount := int64(len(themap))
		b_buf.Write(Int64ToBytes(thecount))
		for key, _ := range themap {
			key_b := []byte(key)
			key_b_len := int64(len(key_b))
			b_buf.Write(Int64ToBytes(key_b_len))
			b_buf.Write(key_b)
			var bs []byte
			bs = []byte(themap[key])
			b_buf.Write(Int64ToBytes(int64(len(bs))))
			b_buf.Write(bs)
		}
	case "map[string]time.Time":
		themap := themaps.(map[string]time.Time)
		thecount := int64(len(themap))
		b_buf.Write(Int64ToBytes(thecount))
		for key, _ := range themap {
			key_b := []byte(key)
			key_b_len := int64(len(key_b))
			b_buf.Write(Int64ToBytes(key_b_len))
			b_buf.Write(key_b)
			var bs []byte
			bs, err = themap[key].MarshalBinary()
			if err != nil {
				return
			}
			b_buf.Write(Int64ToBytes(int64(len(bs))))
			b_buf.Write(bs)
		}
	case "map[string]int64":
		themap := themaps.(map[string]int64)
		thecount := int64(len(themap))
		b_buf.Write(Int64ToBytes(thecount))
		for key, _ := range themap {
			key_b := []byte(key)
			key_b_len := int64(len(key_b))
			b_buf.Write(Int64ToBytes(key_b_len))
			b_buf.Write(key_b)
			var bs []byte
			bs = Int64ToBytes(themap[key])
			b_buf.Write(bs)
		}
	case "map[string]uint64":
		themap := themaps.(map[string]uint64)
		thecount := int64(len(themap))
		b_buf.Write(Int64ToBytes(thecount))
		for key, _ := range themap {
			key_b := []byte(key)
			key_b_len := int64(len(key_b))
			b_buf.Write(Int64ToBytes(key_b_len))
			b_buf.Write(key_b)
			var bs []byte
			bs = Uint64ToBytes(themap[key])
			b_buf.Write(bs)
		}
	}

	b = b_buf.Bytes()
	return
}
