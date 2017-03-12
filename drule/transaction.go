// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/roles"
)

/*
 以下内容为roleio.RolesInOutManager接口的实现
*/

// 读取一个角色
//
// 角色会缓存并配置成写锁被本事务占用，如果在事务州其中不执行StoreRole保存，那么对这个角色的修改也不会被保存，信息将丢失。
func (t *Transaction) ReadRole(id string) (role roles.Roleer, err error) {
	if t.be_delete == true {
		return nil, fmt.Errorf("DRule[Transaction]ReadRole: This transaction has been deleted.")
	}
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_WRITE)
	if err != nil {
		err = fmt.Errorf("DRule[Transacion]ReadRole: %v", err)
		return
	}
	return rolec.role, nil
}

// 写入一个角色
//
// 依然会去缓存中尝试获取角色的写权限，如果找不到,则去写一个新的
func (t *Transaction) StoreRole(role roles.Roleer) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("DRule[Transaction]StoreRole: This transaction has been deleted.")
	}
	t.lock.RLock()
	defer t.lock.RUnlock()
	roleid := role.ReturnId()
	var find bool
	rolec, find := t.tran_cache[roleid]
	if find == true {
		rolec.role = role
	} else {
		rolec, err = t.tran_service.addRole(t.unid, role)
		if err != nil {
			return fmt.Errorf("DRule[Transaction]StoreRole: %v", err)
		}
		t.tran_cache[roleid] = rolec
	}
	return nil
}

// 读一个father
func (t *Transaction) ReadFather(id string) (father string, err error) {
	if t.be_delete == true {
		return "", fmt.Errorf("DRule[Transaction]ReadFather: This transaction has been deleted.")
	}
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
	if t.be_delete == true {
		return fmt.Errorf("This transaction has been deleted.")
	}
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
	if t.be_delete == true {
		return fmt.Errorf("This transaction has been deleted.")
	}
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
	if t.be_delete == true {
		return nil, fmt.Errorf("This transaction has been deleted.")
	}
	rolec, err := t.getrole(id, TRAN_LOCK_MODE_READ)
	if err != nil {
		err = fmt.Errorf("ReadChildren: %v", err)
		return
	}
	children = rolec.role.GetChildren()
	return
}

/*
 以上内容为roleio.RolesInOutManager接口的实现
*/

// 事务执行的处理
func (t *Transaction) Commit() (err error) {
	if t.be_delete == true {
		return fmt.Errorf("This transaction has been deleted.")
	}
	// 构造事务执行的处理信号
	commit_signal := &tranCommitSignal{
		tran_id:       t.unid,
		ask:           TRAN_COMMIT_ASK_COMMIT,
		return_handle: make(chan tranReturnHandle),
	}
	// 发送出去
	t.tran_commit_signal <- commit_signal
	// 开始等返回
	return_sigle := <-commit_signal.return_handle
	//fmt.Println("等到了返回：", t.unid)
	if return_sigle.Status != TRAN_RETURN_HANDLE_OK {
		return return_sigle.Error
	}
	return
}

// 事务的回滚处理
func (t *Transaction) Rollback() (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]Rollback: This transaction has been deleted.")
	}
	// 构造事务执行的处理信号
	rollback_signal := &tranCommitSignal{
		tran_id:       t.unid,
		ask:           TRAN_COMMIT_ASK_ROLLBACK,
		return_handle: make(chan tranReturnHandle),
	}
	// 发送信号
	t.tran_commit_signal <- rollback_signal
	// 开始等返回
	return_signle := <-rollback_signal.return_handle
	if return_signle.Status != TRAN_RETURN_HANDLE_OK {
		return return_signle.Error
	}
	return
}

// 事务的准备。输入将准备的角色ID，让事务可以先尝试获得写权限
func (t *Transaction) Prepare(roleids ...string) (err error) {
	err = t.prepare(roleids)
	if err != nil {
		return fmt.Errorf("drule[Transaction]Prepare: %v", err)
	}
	return
}

func (t *Transaction) prepare(roleids []string) (err error) {
	errall := make([]string, 0)
	for _, oneid := range roleids {
		_, errn := t.getrole(oneid, TRAN_LOCK_MODE_WRITE)
		if errn != nil {
			errall = append(errall, errn.Error())
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
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
