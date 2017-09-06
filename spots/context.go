// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spots

import "bytes"
import "github.com/idcsource/insight00-lib/iendecode"

// 句柄上下文的结构
type Context struct {
	// 上游
	Up map[string]Status
	// 下游
	Down map[string]Status
}

func NewContext() (c Context) {
	c = Context{
		Up:   make(map[string]Status),
		Down: make(map[string]Status),
	}
	return
}

func (c Context) MarshalBinary() (b []byte, err error) {
	up_b, up_lens, err := c.mapByte(c.Up)
	if err != nil {
		return
	}
	down_b, down_lens, err := c.mapByte(c.Down)
	if err != nil {
		return
	}
	lens := up_lens + down_lens + 16
	b = make([]byte, lens)
	copy(b, iendecode.Uint64ToBytes(uint64(up_lens)))
	copy(b[8:], up_b)
	copy(b[up_lens+8:], iendecode.Uint64ToBytes(uint64(down_lens)))
	copy(b[up_lens+16:], down_b)
	return
}

func (c Context) mapByte(m map[string]Status) (b []byte, lens int64, err error) {
	b_buf := bytes.Buffer{}
	for key, _ := range m {
		key_len := iendecode.Uint64ToBytes(uint64(len(key)))
		b_buf.Write(key_len)
		b_buf.Write([]byte(key))
		var s_b []byte
		var s_lens int64
		s_b, err = m[key].MarshalBinary()
		if err != nil {
			return
		}
		s_lens = int64(len(s_b))
		b_buf.Write(iendecode.Uint64ToBytes(uint64(s_lens)))
		b_buf.Write(s_b)
	}
	lens = int64(b_buf.Len())
	b = b_buf.Bytes()
	return
}

func (c *Context) UnmarshalBinary(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	c.Up = make(map[string]Status)
	c.Down = make(map[string]Status)

	up_len := iendecode.BytesToUint64(b[0:8])
	if err != nil {
		return err
	}
	c.Up, err = c.byteMap(b[8 : 8+up_len])
	if err != nil {
		return err
	}
	down_len := iendecode.BytesToUint64(b[8+up_len : 8+up_len+8])
	c.Down, err = c.byteMap(b[16+up_len : 16+up_len+down_len])
	if err != nil {
		return err
	}
	return
}

func (c Context) byteMap(b []byte) (m map[string]Status, err error) {
	m = make(map[string]Status)
	b_buf := bytes.NewBuffer(b)
	var i uint64 = 0
	b_len := uint64(len(b))
	for {
		if i >= b_len {
			break
		}
		key_len_b := b_buf.Next(8)
		key_len := iendecode.BytesToUint64(key_len_b)
		key := b_buf.Next(int(key_len))
		s_len_b := b_buf.Next(8)
		s_len := iendecode.BytesToUint64(s_len_b)
		s_b := b_buf.Next(int(s_len))
		s := Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
		err = s.UnmarshalBinary(s_b)
		if err != nil {
			return
		}
		m[string(key)] = s
		i += 16 + key_len + s_len
	}
	return
}
