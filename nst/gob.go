// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package nst

import (
	"bytes"
	"encoding/gob"
	"reflect"

	"github.com/idcsource/insight00-lib/roles"
)

// 将结构体数据转成Gob再转成[]Byte
func StructGobBytes(e interface{}) ([]byte, error) {
	var gob_buff bytes.Buffer           //建立缓冲
	gob_en := gob.NewEncoder(&gob_buff) //gob开始编码
	err := gob_en.Encode(e)             //gob编码
	if err != nil {
		return nil, err
	}
	gob_b := gob_buff.Bytes() //bytes.buffer转[]byte
	return gob_b, nil
}

// 将[]byte转成Gob再转成结构体
func BytesGobStruct(f_b []byte, stur interface{}) error {
	b_buf := bytes.NewBuffer(f_b) //将[]byte放入bytes的buffer中
	b_go := gob.NewDecoder(b_buf) //将buffer放入gob的decoder中
	err := b_go.Decode(stur)      //将gob解码放入stur
	if err != nil {
		return err
	} else {
		return nil
	}
}

// 将roles.Roleer接口的结构体数据转成Gob再转成[]Byte，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func StructGobBytesForRoleer(p roles.Roleer) ([]byte, error) {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err := enc.Encode(&p)
	if err != nil {
		return nil, err
	}
	return network.Bytes(), nil
}

// 将[]byte转成Gob再转成roles.Roleer接口的结构体，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func BytesGobStructForRoleer(by []byte) (roles.Roleer, error) {
	var p roles.Roleer
	network := bytes.NewBuffer(by)
	dec := gob.NewDecoder(network)
	err := dec.Decode(&p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// 将[]byte转成提供的反射Value
func BytesGobReflect(f_b []byte, v reflect.Value) error {
	b_buf := bytes.NewBuffer(f_b) //将[]byte放入bytes的buffer中
	b_go := gob.NewDecoder(b_buf) //将buffer放入gob的decoder中
	err := b_go.DecodeValue(v)    //将gob解码放入
	if err != nil {
		return err
	} else {
		return nil
	}
}
