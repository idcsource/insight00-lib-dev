// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spots

import "bytes"

import "github.com/idcsource/insight00-lib/iendecode"

type Status struct {
	Int     []int64
	Float   []float64
	Complex []complex128
	String  []string
}

func NewStatus() (s Status) {
	return Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
}

func (s Status) MarshalBinary() (b []byte, err error) {
	var int_b []byte // 80
	int_b, err = iendecode.ToBinary(s.Int)
	if err != nil {
		return
	}
	var float_b []byte // 80
	float_b, err = iendecode.ToBinary(s.Float)
	if err != nil {
		return
	}
	var complex_b []byte // 160
	complex_b, err = iendecode.ToBinary(s.Complex)
	if err != nil {
		return
	}
	string_buf := bytes.Buffer{}
	var string_b_len int64 = 0
	for i := range s.String {
		sb := []byte(s.String[i])
		sb_l := int64(len(sb))
		if sb_l != 0 {
			string_b_len += 8 + sb_l
			string_buf.Write(iendecode.Uint64ToBytes(uint64(sb_l)))
			string_buf.Write(sb)
		} else {
			string_b_len += 8
			string_buf.Write(iendecode.Uint64ToBytes(uint64(sb_l)))
		}
	}
	string_b := string_buf.Bytes()
	lens := 320 + string_b_len
	b = make([]byte, lens)
	copy(b, int_b)
	copy(b[80:], float_b)
	copy(b[160:], complex_b)
	copy(b[320:], string_b)
	return
}

func (s *Status) UnmarshalBinary(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	s.Int = make([]int64, 10)
	s.Float = make([]float64, 10)
	s.Complex = make([]complex128, 10)
	s.String = make([]string, 10)

	err = iendecode.FromBinary(b[0:80], &s.Int)
	if err != nil {
		return
	}
	err = iendecode.FromBinary(b[80:160], &s.Float)
	if err != nil {
		return
	}
	err = iendecode.FromBinary(b[160:320], &s.Complex)
	if err != nil {
		return
	}
	j := 0
	var i uint64 = 320
	for {
		if i >= uint64(len(b)) {
			break
		}
		slen := iendecode.BytesToUint64(b[i : i+8])
		if slen != 0 {
			s.String[j] = string(b[i+8 : i+8+slen])
		}
		i += 8 + slen
		j++
	}
	return
}
