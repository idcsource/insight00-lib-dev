// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package nst

import (
	"crypto/tls"
	"fmt"
	"net"
)

// TCP发送接收数据结构。发送数据的方法必须使用对应类型的接收方法进行接收
type TCP struct {
	tls     bool
	tcp     net.Conn
	tcp_tls *tls.Conn
	buf     int
}

// 新建TCP的发送接收
func NewTCP(tcp net.Conn) *TCP {
	return &TCP{tls: false, tcp: tcp, buf: 1024}
}

func NewTCPtls(tcp *tls.Conn) *TCP {
	return &TCP{tls: true, tcp_tls: tcp, buf: 1024}
}

// 设置缓冲大小
func (t *TCP) SetBuf(buf int) {
	t.buf = buf
}

// 查看缓冲大小
func (t *TCP) GetBuf() int {
	return t.buf
}

// 发送一个结构体（会在方法内部转换成gob），只能用GetStruct()方法接受
func (t *TCP) SendStruct(stru interface{}) (errs error) {
	bytes, errs := StructGobBytes(stru)
	if errs != nil {
		return
	}
	len := len(bytes)
	lens := uint64(len)
	errs = t.SendLen(lens) //发送长度
	if errs != nil {
		return
	}
	errs = t.SendBytes(bytes)
	return
}

// 接收一个结构体（会在方法内部从gob转出），只能接受SendStruct()方法发送的信息
func (t *TCP) GetStruct(stru interface{}) (errs error) {
	lens, errs := t.GetLen()
	if errs != nil {
		return
	}
	bytes, errs := t.GetBytes(lens)
	if errs != nil {
		return
	}
	BytesGobStruct(bytes, stru)
	return
}

// 发送一串数据流（包括字节流的长度），必须用GetData()方法接收
func (t *TCP) SendData(bytes []byte) (errs error) {
	len := len(bytes)
	lens := uint64(len)
	errs = t.SendLen(lens) //发送长度
	if errs != nil {
		return
	}
	errs = t.SendBytes(bytes)
	return
}

// 接收一串数据流（包括字节流的长度），特定接收SendData()方法发送的数据
func (t *TCP) GetData() (bytes []byte, errs error) {
	lens, errs := t.GetLen()
	if errs != nil {
		return nil, errs
	}
	bytes, errs = t.GetBytes(lens)
	return
}

// 发送一个长度属性，也就是发送uint64
func (t *TCP) SendLen(len uint64) (errs error) {
	vb := Uint64ToBytes(len)
	errs = t.SendBytes(vb)
	return
}

// 接收一个长度属性，也就是接收uint64
func (t *TCP) GetLen() (len uint64, errs error) {
	bytes, errs := t.GetBytes(8)
	if errs != nil {
		return 0, errs
	}
	len = BytesToUint64(bytes)
	return
}

// 发送一个流程版本号，也可以发送流程编号，就是发送一个uint32
func (t *TCP) SendVer(version uint32) (errs error) {
	vb := Uint32ToBytes(version)
	errs = t.SendBytes(vb)
	return
}

// 接收一个流程版本号，也就是接收一个uint32
func (t *TCP) GetVer() (version uint32, errs error) {
	bytes, errs := t.GetBytes(4)
	version = BytesToUint32(bytes)
	return
}

// 发送状态，也就是发送uint8
func (t *TCP) SendStat(status uint8) (errs error) {
	vb := Uint8ToBytes(status)
	errs = t.SendBytes(vb)
	return
}

// 接收状态，也就是接收uint8
func (t *TCP) GetStat() (status uint8, errs error) {
	bytes, errs := t.GetBytes(1)
	if errs != nil {
		return
	}
	status = BytesToUint8(bytes)
	return
}

// 发送字节切片（没有字节长度信息）
func (t *TCP) SendBytes(bytes []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("nst[TCP]SendBytes: %v", e)
		}
	}()
	if t.tls == false {
		_, err = t.tcp.Write(bytes)
	} else {
		_, err = t.tcp_tls.Write(bytes)
	}
	if err != nil {
		return
	}
	return
}

// 接收字节切片（没有字节长度信息）
func (t *TCP) GetBytes(len uint64) (returnByte []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("nst[TCP]GetBytes: %v", e)
		}
	}()
	returnByte = make([]byte, 0, len)
	for {
		var err error
		tempdata := []byte{}
		if len < uint64(t.buf) {
			tempdata = make([]byte, len)
		} else {
			tempdata = make([]byte, t.buf)
		}
		var r int
		if t.tls == false {
			r, err = t.tcp.Read(tempdata)
		} else {
			r, err = t.tcp_tls.Read(tempdata)
		}
		if err != nil {
			return returnByte, err
		}
		returnByte = append(returnByte, tempdata[:r]...)

		len = len - uint64(r)

		if len == 0 {
			break
		}
	}
	return returnByte, err
}

// 设置长连接模式
//func (t *TCP) SetKeepAlive(keepalive bool) error {
//	return t.tcp.SetKeepAlive(keepalive)
//}

// 关闭连接
func (t *TCP) Close() (err error) {
	if t.tls == true {
		err = t.tcp_tls.Close()
		if err != nil {
			return
		}
	}
	err = t.tcp.Close()
	return
}
