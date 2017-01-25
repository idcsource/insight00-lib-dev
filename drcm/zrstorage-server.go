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

// ExecTCP nst的ConnExecer接口
func (z *ZrStorage) ExecTCP(conn_exec *nst.ConnExec) (err error) {
	// 接收前导
	prefix_stat_b, err := conn_exec.GetData()
	if err != nil {
		z.logerr(err)
		return err
	}
	// 解码前导
	prefix_stat := Net_PrefixStat{}
	err = nst.BytesGobStruct(prefix_stat_b, &prefix_stat)
	if err != nil {
		z.logerr(err)
		// 发送关闭
		conn_exec.SendClose()
		return err
	}
	// 查看身份验证码
	if prefix_stat.Code != z.code {
		// 如果身份码不相符，发送关闭
		conn_exec.SendClose()
		return nil
	}
	// 遍历所有操作,转到相应方法
	switch prefix_stat.Operate {

	case OPERATE_TOSTORE:

	case OPERATE_READ_ROLE:

	case OPERATE_WRITE_ROLE:

	case OPERATE_NEW_ROLE:

	case OPERATE_DEL_ROLE:

	case OPERATE_GET_DATA:

	case OPERATE_SET_DATA:

	case OPERATE_SET_FATHER:

	case OPERATE_GET_FATHER:

	case OPERATE_RESET_FATHER:

	case OPERATE_SET_CHILDREN:

	case OPERATE_GET_CHILDREN:

	case OPERATE_RESET_CHILDREN:

	case OPERATE_ADD_CHILD:

	case OPERATE_DEL_CHILD:

	case OPERATE_EXIST_CHILD:

	case OPERATE_SET_FRIENDS:

	case OPERATE_GET_FRIENDS:

	case OPERATE_RESET_FRIENDS:

	case OPERATE_ADD_FRIEND:

	case OPERATE_DEL_FRIEND:

	case OPERATE_CHANGE_FRIEND:

	case OPERATE_SAME_BIND_FRIEND:

	case OPERATE_ADD_CONTEXT:

	case OPERATE_DROP_CONTEXT:

	case OPERATE_GET_CONTEXTS_NAME:

	case OPERATE_READ_CONTEXT:

	case OPERATE_SAME_BIND_CONTEXT:

	case OPERATE_ADD_CONTEXT_BIND:

	case OPERATE_DEL_CONTEXT_BIND:

	case OPERATE_CHANGE_CONTEXT_BIND:

	case OPERATE_CONTEXT_SAME_BIND:

	case OPERATE_ADD_CONTEXT_UP:

	case OPERATE_DEL_CONTEXT_UP:

	case OPERATE_CHANGE_CONTEXT_UP:

	case OPERATE_SAME_BIND_CONTEXT_UP:

	case OPERATE_ADD_CONTEXT_DOWN:

	case OPERATE_DEL_CONTEXT_DOWN:

	case OPERATE_CHANGE_CONTEXT_DOWN:

	case OPERATE_SAME_BIND_CONTEXT_DOWN:

	case OPERATE_SET_FRIEND_STATUS:

	case OPERATE_GET_FRIEND_STATUS:

	case OPERATE_SET_CONTEXT_STATUS:

	case OPERATE_GET_CONTEXT_STATUS:

	case OPERATE_SET_CONTEXTS:

	case OPERATE_GET_CONTEXTS:

	case OPERATE_RESET_CONTEXTS:

	default:
		conn_exec.SendClose()
		return nil
	}
	return nil
}
