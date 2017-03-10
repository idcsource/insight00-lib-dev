// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import "fmt"

// 读一个father
func (t *Transaction) ReadFather(id string) (father string, err error) {
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_READ)
	if err != nil {
		err = fmt.Errorf("ReadFather: %v", err)
		return
	}
	father = rolec.role.GetFather()
	return
}

// 写一个father
func (t *Transaction) WriteFather(id, father string) (err error) {
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_WRITE)
	if err != nil {
		err = fmt.Errorf("WriteFather: %v", err)
		return
	}
	rolec.role.SetFather(father)
	return
}

// 写一个child
func (t *Transaction) WriteChild(id, child string) (err error) {
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_WRITE)
	if err != nil {
		err = fmt.Errorf("WriteChild: %v", err)
		return
	}
	rolec.role.AddChild(child)
	return
}

// 获取children
func (t *Transaction) ReadChildren(id string) (children []string, err error) {
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_READ)
	if err != nil {
		err = fmt.Errorf("ReadChildren: %v", err)
		return
	}
	children = rolec.role.GetChildren()
	return
}

// 事务执行的处理
func (t *Transaction) Commit() (err error) {
	// 构造事务执行的处理信号
	commit_signal := &tranCommitSignal{
		tran_id:       t.unid,
		ask:           TRAN_COMMIT_ASK_COMMIT,
		return_handle: make(chan tranReturnHandle),
	}
	// 发送出去
	t.tran_commit_signal <- commit_signal
	// 开始看返回
	return_sigle := <-commit_signal.return_handle
	if return_sigle.Status != TRAN_RETURN_HANDLE_OK {
		return return_sigle.Error
	}
	return
}

func (t *Transaction) getrole(id string, lockmode uint8) (rolec *roleCache, err error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	var find bool
	rolec, find = t.tran_cache[id]
	if find == true {
		return
	} else {
		rolec, err = t.tran_service.getRole(t.unid, id, lockmode)
		if err == nil && lockmode == TRAN_LOCK_MODE_WRITE {
			t.tran_cache[id] = rolec
		}
		return
	}
}
