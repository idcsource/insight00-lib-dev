// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 向slave发送前导状态，也就是身份验证码和要操作的状态，并获取slave是否可以继续传输的要求
func SendPrefixStat(process *nst.ProgressData, code string, operate int) (receipt Net_SlaveReceipt, err error) {
	thestat := Net_PrefixStat{
		Operate: operate,
		Code:    code,
	}
	statbyte, err := nst.StructGobBytes(thestat)
	if err != nil {
		return
	}
	rdata, err := process.SendAndReturn(statbyte)
	if err != nil {
		return
	}
	receipt = Net_SlaveReceipt{}
	err = nst.BytesGobStruct(rdata, &receipt)
	return
}

// 从[]byte解码SlaveReceipt
func DecodeSlaveReceipt(b []byte) (receipt Net_SlaveReceipt, err error) {
	receipt = Net_SlaveReceipt{}
	err = nst.BytesGobStruct(b, &receipt)
	return
}

// 从[]byte解码SlaveReceipt带数据体
func DecodeSlaveReceiptData(b []byte) (receipt Net_SlaveReceipt_Data, err error) {
	receipt = Net_SlaveReceipt_Data{}
	err = nst.BytesGobStruct(b, &receipt)
	return
}

// 发送数据并解码返回的SlaveReceipt
func SendAndDecodeSlaveReceipt(cprocess *nst.ProgressData, data []byte) (receipt Net_SlaveReceipt, err error) {
	s_r_b, err := cprocess.SendAndReturn(data)
	if err != nil {
		return
	}
	receipt, err = DecodeSlaveReceipt(s_r_b)
	return
}

// 发送数据并解码返回的SlaveReceipt_Data
func SendAndDecodeSlaveReceiptData(cprocess *nst.ProgressData, data []byte) (receipt Net_SlaveReceipt_Data, err error) {
	s_r_b, err := cprocess.SendAndReturn(data)
	if err != nil {
		return
	}
	receipt, err = DecodeSlaveReceiptData(s_r_b)
	return
}
