// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 硬存储（永久存储），一套文件型存储数据库。
//
// 实现RolesInOutManager的接口（依靠roles中的NilReadWrite，并非全部实现），
// 对角色的信息与关系进行永久存储。
//
// 需要提供*cpool.Block类型的配置信息。
// 目前必需的配置信息配置项为：
// 		[local]
// 			path = one_path_name		# 存储数据库的保存位置
// 			path_deep = 2			# 数据库结构的路径层级，建议1或2就可以了
// TODO：分布式存储，自管锁
package hardstore

import (
	"fmt"
	"os"
	"errors"
	"io/ioutil"
	
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 存储器类型
type HardStore struct {
	*rolesio.NilReadWrite
	config				*cpool.Block
	local_path			string
	path_deep			int64
	relation_name		string
	context_name		string
}

type roleRelation struct {
	Father string														// 父角色（拓扑结构层面）
	Children []string													// 虚拟的子角色群，只保存键名
	Friends map[string]roles.Status										// 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	Contexts map[string]roles.Context
}

// 新建一个存储实例，如果配置文件缺失必须的配置项或配置项中指定路径无法操作都将返回错误
func NewHardStore (config *cpool.Block) (*HardStore, error) {
	var local_path string;
	var path_deep int64;
	var err error;
	local_path, err = config.GetConfig("local.path");
	if err != nil {
		return nil, errors.New("hardstore: NewHardStore: The configure have not local_path !");
	}
	local_path = pubfunc.LocalPath(local_path);
	
	path_deep, err = config.TranInt64("local.path_deep");
	if err != nil {
		return nil, errors.New("hardstore: NewHardStore: The configure have not local_deep !");
	}
	
	path_info, err := os.Stat(local_path);
	if err != nil {
		return nil, errors.New("hardstore: NewHardStore: The loal_path have not be access !");
	}
	if path_info.IsDir() != true {
		return nil, errors.New("hardstore: NewHardStore: The loal_path have not a path !");
	}
	
	l_path_name := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j" , "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"};
	l_path := make([]string,0);
	
	deployed_file := local_path + "deployed";
	if pubfunc.FileExist(deployed_file) == false {
	
		for i:= 0; i< int(path_deep); i++ {
			if len(l_path) == 0 {
				for _, v := range l_path_name {
					l_path = append(l_path, local_path + v + "/");
				}
			} else {
				ll_path := make([]string,0);
				for _, v := range l_path {
					for _, v2 := range l_path_name {
						ll_path = append(ll_path, v + v2 + "/");
					}
				}
				l_path = append(l_path, ll_path...);
			}
		}
		for _, v := range l_path {
			if pubfunc.FileExist(v) == true {
				continue;
			} else {
				err4 := os.Mkdir(v, 0700);
				if err4 != nil {
					return nil, fmt.Errorf("hardstore: %v", err4);
				}
			}
		}
		f_byte := []byte("Have Deployed");
		ioutil.WriteFile(deployed_file, f_byte, 0600);
	
	}

	hardstore := &HardStore{
		NilReadWrite: rolesio.NewNilReadWrite(),
		config: config,
		local_path: local_path,
		path_deep: path_deep,
		relation_name: "_relation",
		context_name: "_context",
	};
	return hardstore, nil;
}

// 根据角色的Id找到存放路径
func (h *HardStore) findRoleFilePath (name string) string {
	path := h.local_path;
	for i:=0; i<int(h.path_deep); i++ {
		path = path + string(name[i]) + "/";
	}
	return path;
}

// 从存储中读取一个角色的本体，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func (h *HardStore) ReadRole (id string) (roles.Roleer, error) {
	path := h.findRoleFilePath(id);
	f_name := path + id;
	f_relation_name := f_name + h.relation_name;
	if pubfunc.FileExist(f_name) == false || pubfunc.FileExist(f_relation_name) == false {
		return nil, errors.New("hardstore: ReadRole: Can't find the Role " + id );
	}
	
	r_byte, err := ioutil.ReadFile(f_name);
	if err != nil {
		return nil, fmt.Errorf("hardstore: ReadRole: %v", err);
	}
	r_r_byte, err2 := ioutil.ReadFile(f_relation_name);
	if err2 != nil {
		return nil, fmt.Errorf("hardstore: ReadRole: %v", err2);
	}
	
	var r_role roles.Roleer;
	r_role, err3 := nst.BytesGobStructForRoleer(r_byte);
	if err3 != nil {
		return nil, fmt.Errorf("hardstore: ReadRole: %v", err3);
	}
	var r_ralation roleRelation;
	err4 := nst.BytesGobStruct(r_r_byte, &r_ralation);
	if err4 != nil {
		return nil, fmt.Errorf("hardstore: ReadRole: %v", err4);
	}
	
	r_role.SetFather(r_ralation.Father);
	r_role.SetChildren(r_ralation.Children);
	r_role.SetFriends(r_ralation.Friends);
	r_role.SetContexts(r_ralation.Contexts);
	return r_role, nil;
}

// 写入一个角色的本体到存储，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func (h *HardStore) StoreRole (role roles.Roleer) error {
	id := role.ReturnId();
	self_change := role.ReturnChanged(roles.SELF_CHANGED);
	data_change := role.ReturnChanged(roles.DATA_CHANGED);
	
	path := h.findRoleFilePath(id);
	
	f_name := path + id;
	f_ralation_name := f_name + h.relation_name;
	
	if pubfunc.FileExist(f_name) == true && pubfunc.FileExist(f_ralation_name) == true{
		if self_change == false && data_change == false {
			return nil;
		}
	} 
	
	var r_byte []byte;
	if data_change == true {
		var err error;
		r_byte, err = nst.StructGobBytesForRoleer(role);
		if err != nil {
			return fmt.Errorf("hardstore: StoreRole: %v", err);
		}
	}
	
	var r_ralation_byte []byte;
	if self_change == true || pubfunc.FileExist(f_ralation_name) == false {
		r_ralation := roleRelation{
			Father: role.GetFather(),
			Children: role.GetChildren(),
			Friends: role.GetFriends(),
			Contexts: role.GetContexts(),
		};
		var err2 error;
		r_ralation_byte, err2 = nst.StructGobBytes(r_ralation);
		if err2 != nil {
			return fmt.Errorf("hardstore: StoreRole: %v", err2);
		}
	}
	if data_change == true || pubfunc.FileExist(f_name) == false {
		err3 := ioutil.WriteFile(f_name, r_byte, 0600);
		if err3 != nil{
			return fmt.Errorf("hardstore: StoreRole: %v", err3);
		}
	}
	if self_change == true || pubfunc.FileExist(f_ralation_name) == false {
		err4 := ioutil.WriteFile(f_ralation_name, r_ralation_byte, 0600);
		if err4 != nil{
			return fmt.Errorf("hardstore: StoreRole: %v", err4);
		}
	}
	return nil;
}

// 删除掉名为name的角色
func (h *HardStore) DeleteRole (name string) (err error) {
	path := h.findRoleFilePath(name);
	f_name := path + name;
	f_ralation_name := f_name + h.relation_name;
	if pubfunc.FileExist(f_name) == true {
		os.Remove(f_name);
	}
	if pubfunc.FileExist(f_ralation_name) == true {
		os.Remove(f_ralation_name);
	}
	return;
}
