// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"fmt"
)

type DRuleError struct {
	Code DRuleReturnStatus
	Err  error
}

func NewDRuleError() (err DRuleError) {
	return DRuleError{
		Code: DATA_NO_RETRUN,
		Err:  nil,
	}
}

// 返回错误
func (errs *DRuleError) IsError() (err error) {
	if errs.Err == nil || len(errs.Err.Error()) == 0 {
		if errs.Code == DATA_ALL_OK {
			err = nil
			return
		} else {
			err = fmt.Errorf(errs.CodeString())
			return
		}
	} else {
		err = errs.Err
		return
	}
}

// 返回错误的字符串
func (errs *DRuleError) String() (s string) {
	if errs.Err != nil {
		s = errs.Err.Error()
	}
	return
}

// 返回状态码
func (errs *DRuleError) StatCode() (code DRuleReturnStatus) {
	return errs.Code
}

// 返回状态的字符串
func (errs *DRuleError) CodeString() (s string) {
	switch errs.Code {
	case DATA_NO_RETRUN:
		s = "Data no retrun"
	case DATA_NOT_EXPECT:
		s = "Data not expect"
	case DATA_ALL_OK:
		s = "Data all ok"
	case DATA_END:
		s = "Data end"
	case DATA_PLEASE:
		s = "Data please"
	case DATA_WILL_SEND:
		s = "Data will send"
	case DATA_RETURN_ERROR:
		s = "Data return error"
	case DATA_RETURN_IS_TRUE:
		s = "Data return is true"
	case DATA_RETURN_IS_FALSE:
		s = "Data return is false"
	case DATA_TRAN_NOT_EXIST:
		s = "Data transaction not exist"
	case DATA_DRULE_CLOSED:
		s = "DRule closed"
	case DATA_DRULE_NOT_PAUSED:
		s = "DRule not paused."
	case DATA_USER_NOT_LOGIN:
		s = "User not login"
	case DATA_USER_EXIST:
		s = "User already exist"
	case DATA_USER_NO_EXIST:
		s = "User not exist"
	case DATA_USER_NO_AUTHORITY:
		s = "User not have authority"
	case DATA_AREA_EXIST:
		s = "Area exist."
	case DATA_AREA_NO_EXIST:
		s = "Area not exist."
	case DATA_DRULE_OPERATOR_EXIST:
		s = "DRule remote perator exist."
	case DATA_DRULE_OPERATOR_NO_EXIST:
		s = "DRule remote operator not exist."
	default:
		s = "unkown"
	}
	return
}
