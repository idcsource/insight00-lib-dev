// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package rolesio

import (
	"github.com/idcsource/Insight-0-0-lib/roles"

	"errors"
)

// NilReadWrite 为Roles的空输入输出方法
// 只是单纯的实现了RolesInOutManager接口的所有方法，但并没有任何实际用处
// 提供此方法目的在于当不需要Roles拥有读取写入永久应存储时的占位
// 忽略所有的内容，并不会弹出任何错误
// 用户自定义的硬存储方法可以直接继承NilReadWrite
type NilReadWrite struct {
	err error
}

func NewNilReadWrite() *NilReadWrite {
	return &NilReadWrite{
		err: errors.New("This is NilReadWrite, it don't opreate any thing."),
	}
}

func (io *NilReadWrite) ReadRole(name string) (roles.Roleer, error) {
	return nil, io.err
}

func (io *NilReadWrite) StoreRole(role roles.Roleer) error {
	return io.err
}

func (io *NilReadWrite) DeleteRole(name string) error {
	return io.err
}

func (io *NilReadWrite) WriteFather(id, father string) error {
	return io.err
}

func (io *NilReadWrite) ReadFather(id string) (father string, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ResetFather(id string) error {
	return io.err
}

func (io *NilReadWrite) ReadChildren(name string) ([]string, error) {
	return nil, io.err
}

func (io *NilReadWrite) WriteChildren(name string, children []string) error {
	return io.err
}

func (io *NilReadWrite) ResetChildren(name string) error {
	return io.err
}

func (io *NilReadWrite) WriteChild(name, child string) error {
	return io.err
}

func (io *NilReadWrite) DeleteChild(name, child string) error {
	return io.err
}

func (io *NilReadWrite) ExistChild(id, child string) (have bool, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ReadSameBindFriendsId(id string, bind int64) (roles []string, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ReadFriends(name string) (binds map[string]roles.Status, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) WriteFriends(name string, friends map[string]roles.Status) error {
	return io.err
}

func (io *NilReadWrite) ResetFriends(name string) error {
	return io.err
}

func (io *NilReadWrite) WriteFriend(name, friend string, relationship int64) error {
	return io.err
}

func (io *NilReadWrite) DeleteFriend(name, friend string) error {
	return io.err
}

func (io *NilReadWrite) ExistFriend(id, friend string) (bind int64, have bool, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) CreateContext(id, contextname string) error {
	return io.err
}

func (io *NilReadWrite) DeleteContextBind(id, contextname string, upordown uint8, upname string) error {
	return io.err
}

func (io *NilReadWrite) DropContext(id, contextname string) error {
	return io.err
}

/*func (io *NilReadWrite) ExistContext (id, contextname string) (have bool, err error) {
	err = io.err;
	return;
}*/

func (io *NilReadWrite) ReadContext(id, contextname string) (context roles.Context, have bool, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ReadContextSameBind(id, contextname string, upordown uint8, bind int64) (rolesid []string, have bool, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ReadContextsName(id string) (names []string, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) WriteFriendStatus(id, friends string, bindbit int, value interface{}) (err error) {
	return io.err
}

func (io *NilReadWrite) ReadFriendStatus(id, friends string, bindbit int, value interface{}) (err error) {
	return io.err
}

func (io *NilReadWrite) WriteContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	return io.err
}

func (io *NilReadWrite) ReadContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	return io.err
}

func (io *NilReadWrite) WriteContexts(id string, context map[string]roles.Context) error {
	return io.err
}

func (io *NilReadWrite) ReadContexts(id string) (contexts map[string]roles.Context, err error) {
	err = io.err
	return
}

func (io *NilReadWrite) ResetContexts(id string) error {
	return io.err
}

func (io *NilReadWrite) WriteData(id, name string, data interface{}) (err error) {
	return io.err
}

func (io *NilReadWrite) ReadData(id, name string, data interface{}) (err error) {
	err = io.err
	return
}
