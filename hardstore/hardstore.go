// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// HardStore 硬存储（永久存储），一套文件型存储数据库。
//
// 实现RolesInOutManager的接口（依靠roles中的NilReadWrite，并非全部实现），
// 对角色的信息与关系进行永久存储。
//
// Change:现在你可以使角色的ID有意义了，不用关注存储时的文件名。
//
// 需要提供*cpool.Section类型的配置信息。
// 目前必需的配置信息配置项为：
// 			path = one_path_name		# 存储数据库的保存位置
// 			path_deep = 2			# 数据库结构的路径层级，建议1或2就可以了
//
// 分布式存储和锁机制已经在drcm包的ZrStorage和rcontrol包中得到实现。HardStore仅作为这两个包的底层存储使用。
package hardstore

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 新建一个存储实例，如果配置文件缺失必须的配置项或配置项中指定路径无法操作都将返回错误
func NewHardStore(config *cpool.Section) (*HardStore, error) {
	var local_path string
	var path_deep int64
	var err error
	local_path, err = config.GetConfig("path")
	if err != nil {
		return nil, errors.New("hardstore[HardStore]NewHardStore: The configure have not local_path !")
	}
	local_path = pubfunc.LocalPath(local_path)

	path_deep, err = config.TranInt64("path_deep")
	if err != nil {
		return nil, errors.New("hardstore[HardStore]NewHardStore: The configure have not local_deep !")
	}

	path_info, err := os.Stat(local_path)
	if err != nil {
		return nil, errors.New("hardstore[HardStore]NewHardStore: The loal_path have not be access !")
	}
	if path_info.IsDir() != true {
		return nil, errors.New("hardstore[HardStore]NewHardStore: The loal_path have not a path !")
	}

	l_path_name := []string{"a", "b", "c", "d", "e", "f", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	l_path := make([]string, 0)

	deployed_file := local_path + "deployed"
	if pubfunc.FileExist(deployed_file) == false {

		for i := 0; i < int(path_deep); i++ {
			if len(l_path) == 0 {
				for _, v := range l_path_name {
					l_path = append(l_path, local_path+v+"/")
				}
			} else {
				ll_path := make([]string, 0)
				for _, v := range l_path {
					for _, v2 := range l_path_name {
						ll_path = append(ll_path, v+v2+"/")
					}
				}
				l_path = append(l_path, ll_path...)
			}
		}
		for _, v := range l_path {
			if pubfunc.FileExist(v) == true {
				continue
			} else {
				err4 := os.Mkdir(v, 0700)
				if err4 != nil {
					return nil, fmt.Errorf("hardstore[HardStore]NewHardStore: %v", err4)
				}
			}
		}
		f_byte := []byte("Have Deployed")
		ioutil.WriteFile(deployed_file, f_byte, 0600)

	}

	hardstore := &HardStore{
		NilReadWrite:  rolesio.NewNilReadWrite(),
		config:        config,
		local_path:    local_path,
		path_deep:     path_deep,
		version_name:  "_version",
		relation_name: "_relation",
		body_name:     "_body",
		data_name:     "_data",
	}
	return hardstore, nil
}

// 根据角色的Id找到存放路径
func (h *HardStore) findRoleFilePath(name string) string {
	path := h.local_path
	for i := 0; i < int(h.path_deep); i++ {
		path = path + string(name[i]) + "/"
	}
	return path
}

// 角色是否存在，只是检查主文件是否存在，需要使用Middle保存的才游泳
func (h *HardStore) RoleExist(id string) (have bool) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	return pubfunc.FileExist(f_name)
}

// 从存储中读取一个角色的本体，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func (h *HardStore) ReadRole(id string) (roles.Roleer, error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_relation_name := f_name + h.relation_name
	f_version_name := f_name + h.version_name
	f_body_name := f_name + h.body_name
	if pubfunc.FileExist(f_body_name) == false || pubfunc.FileExist(f_relation_name) == false || pubfunc.FileExist(f_version_name) == false {
		return nil, errors.New("hardstore[HardStore]ReadRole: Can't find the Role " + id)
	}

	r_byte, err := ioutil.ReadFile(f_body_name)
	if err != nil {
		return nil, fmt.Errorf("hardstore[HardStore]ReadRole: %v", err)
	}
	r_r_byte, err2 := ioutil.ReadFile(f_relation_name)
	if err2 != nil {
		return nil, fmt.Errorf("hardstore[HardStore]ReadRole: %v", err2)
	}
	r_v_byte, err2_0 := ioutil.ReadFile(f_version_name)
	if err2_0 != nil {
		return nil, fmt.Errorf("hardstore[HardStore]ReadRole: %v", err2_0)
	}

	role, err := h.DecodeRole(r_byte, r_r_byte, r_v_byte)

	return role, err
}

// 读取一个Middle格式
func (h *HardStore) ReadMiddle(id string) (mid RoleMiddleData, err error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name

	if pubfunc.FileExist(f_name) == false {
		err = fmt.Errorf("hardstore[HardStore]ReadMiddle: Can't find the Role " + id)
		return
	}
	f_version_b, err := ioutil.ReadFile(f_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_version := RoleVersion{}
	err = nst.BytesGobStruct(f_version_b, &f_version)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_relation_b, err := ioutil.ReadFile(f_ralation_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_relation := RoleRelation{}
	err = nst.BytesGobStruct(f_relation_b, &f_relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_normal_b, err := ioutil.ReadFile(f_data_name + "_normal")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_normal := RoleDataNormal{}
	err = nst.BytesGobStruct(f_data_normal_b, &f_data_normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_slice_b, err := ioutil.ReadFile(f_data_name + "_slice")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_slice := RoleDataSlice{}
	err = nst.BytesGobStruct(f_data_slice_b, &f_data_slice)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_stringmap_b, err := ioutil.ReadFile(f_data_name + "_stringmap")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_stringmap := RoleDataStringMap{}
	err = nst.BytesGobStruct(f_data_stringmap_b, &f_data_stringmap)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	// 合成中间数据
	mid = RoleMiddleData{
		Version:   f_version,
		Relation:  f_relation,
		Normal:    f_data_normal,
		Slice:     f_data_slice,
		StringMap: f_data_stringmap,
	}
	return
}

// 读出一个角色
func (h *HardStore) ReadRoleByMiddle(id string, role roles.Roleer) (err error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name

	if pubfunc.FileExist(f_name) == false {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: Can't find the Role " + id)
		return
	}

	f_version_b, err := ioutil.ReadFile(f_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_version := RoleVersion{}
	err = nst.BytesGobStruct(f_version_b, &f_version)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_relation_b, err := ioutil.ReadFile(f_ralation_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_relation := RoleRelation{}
	err = nst.BytesGobStruct(f_relation_b, &f_relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_normal_b, err := ioutil.ReadFile(f_data_name + "_normal")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_normal := RoleDataNormal{}
	err = nst.BytesGobStruct(f_data_normal_b, &f_data_normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_slice_b, err := ioutil.ReadFile(f_data_name + "_slice")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_slice := RoleDataSlice{}
	err = nst.BytesGobStruct(f_data_slice_b, &f_data_slice)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_stringmap_b, err := ioutil.ReadFile(f_data_name + "_stringmap")
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_stringmap := RoleDataStringMap{}
	err = nst.BytesGobStruct(f_data_stringmap_b, &f_data_stringmap)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	// 合成中间数据
	mid := RoleMiddleData{
		Version:   f_version,
		Relation:  f_relation,
		Normal:    f_data_normal,
		Slice:     f_data_slice,
		StringMap: f_data_stringmap,
	}
	// 解码
	err = h.DecodeMiddleToRole(mid, role)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
	}
	return
}

// 解码一个角色，将二进制的角色存储进行解码
func (h *HardStore) DecodeRole(roleb, relab, verb []byte) (role roles.Roleer, err error) {
	return DecodeRole(roleb, relab, verb)
}

func (h *HardStore) DecodeMiddleToRole(mid RoleMiddleData, role roles.Roleer) (err error) {
	return DecodeMiddleToRole(mid, role)
}

// 存储角色，直接从[]byte结构
func (h *HardStore) StoreRoleByMiddleByte(b []byte) (err error) {
	mid := RoleMiddleData{}
	err = nst.BytesGobStruct(b, &mid)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	id := mid.Version.Id
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name

	// 存数据
	normal_b, err := nst.StructGobBytes(mid.Normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	slice_b, err := nst.StructGobBytes(mid.Slice)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	stringmap_b, err := nst.StructGobBytes(mid.StringMap)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_normal", normal_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_slice", slice_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_stringmap", stringmap_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}

	// 保存关系和Version
	version_b, err := nst.StructGobBytes(mid.Version)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	relation_b, err := nst.StructGobBytes(mid.Relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_ralation_name, relation_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_name, version_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	return
}

// 存储角色，保存为中间数据格式
func (h *HardStore) StoreRoleByMiddle(role roles.Roleer) (err error) {
	id := role.ReturnId()
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	// f_name用来保存version了
	// f_version_name := f_name + h.version_name
	f_data_name := f_name + h.data_name

	// 编码成为中间格式
	mid, err := h.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}

	// 存数据
	normal_b, err := nst.StructGobBytes(mid.Normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	slice_b, err := nst.StructGobBytes(mid.Slice)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	stringmap_b, err := nst.StructGobBytes(mid.StringMap)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_normal", normal_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_slice", slice_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name+"_stringmap", stringmap_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}

	// 保存关系和Version
	version_b, err := nst.StructGobBytes(mid.Version)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	relation_b, err := nst.StructGobBytes(mid.Relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_ralation_name, relation_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_name, version_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	return
}

// 写入一个角色的本体到存储，需要提前用encoding/gob包中的Register方法注册符合roles.Roleer接口的数据类型。
func (h *HardStore) StoreRole(role roles.Roleer) (err error) {
	id := role.ReturnId()
	hashid := random.GetSha1Sum(id)
	//self_change := role.ReturnChanged(roles.SELF_CHANGED)
	//data_change := role.ReturnChanged(roles.DATA_CHANGED)

	path := h.findRoleFilePath(hashid)

	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	f_body_name := f_name + h.body_name
	f_version_name := f_name + h.version_name

	body, ralation, version, err := h.EncodeRole(role)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f_body_name, body, 0600)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f_ralation_name, ralation, 0600)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f_version_name, version, 0600)
	if err != nil {
		return err
	}

	/*
		if pubfunc.FileExist(f_body_name) == true && pubfunc.FileExist(f_ralation_name) == true && pubfunc.FileExist(f_version_name) == true {
			if self_change == false && data_change == false {
				return nil
			}
		}


		if pubfunc.FileExist(f_version_name) == false {
			version := role.Version()
			role_version := roleVersion{
				Version: version,
			}
			role_version_b, err := nst.StructGobBytes(role_version)
			if err != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err)
			}
			err = ioutil.WriteFile(f_version_name, role_version_b, 0600)
			if err != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err)
			}
		}

		var r_byte []byte
		if data_change == true {
			var err error
			r_byte, err = nst.StructGobBytesForRoleer(role)
			if err != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err)
			}
		}

		var r_ralation_byte []byte
		if self_change == true || pubfunc.FileExist(f_ralation_name) == false {
			r_ralation := roleRelation{
				Father:   role.GetFather(),
				Children: role.GetChildren(),
				Friends:  role.GetFriends(),
				Contexts: role.GetContexts(),
			}
			var err2 error
			r_ralation_byte, err2 = nst.StructGobBytes(r_ralation)
			if err2 != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err2)
			}
		}
		if data_change == true || pubfunc.FileExist(f_name) == false {
			err3 := ioutil.WriteFile(f_body_name, r_byte, 0600)
			if err3 != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err3)
			}
		}
		if self_change == true || pubfunc.FileExist(f_ralation_name) == false {
			err4 := ioutil.WriteFile(f_ralation_name, r_ralation_byte, 0600)
			if err4 != nil {
				return fmt.Errorf("hardstore: StoreRole: %v", err4)
			}
		}
	*/
	return nil
}

// 编码角色，将角色编码为两个部分的[]byte，一个是角色本身的数据roleb，一个是角色的关系relab
func (h *HardStore) EncodeRole(role roles.Roleer) (roleb, relab, verb []byte, err error) {
	return EncodeRole(role)
}

func (h *HardStore) EncodeRoleToMiddle(role roles.Roleer) (mid RoleMiddleData, err error) {
	return EncodeRoleToMiddle(role)
}

// 删除掉名为name的角色
func (h *HardStore) DeleteRole(id string) (err error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_body_name := f_name + h.body_name
	f_ralation_name := f_name + h.relation_name
	f_version_name := f_name + h.version_name
	if pubfunc.FileExist(f_body_name) == true {
		os.Remove(f_body_name)
	}
	if pubfunc.FileExist(f_ralation_name) == true {
		os.Remove(f_ralation_name)
	}
	if pubfunc.FileExist(f_version_name) == true {
		os.Remove(f_version_name)
	}
	return
}

// 中间类型的存储数据
func (h *HardStore) WriteDataByMiddle(id, name string, datas interface{}) (err error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	//f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name

	if pubfunc.FileExist(f_name) == false {
		err = fmt.Errorf("hardstore[HardStore]WriteDataByMiddle: Can't find the Role " + id)
		return
	}

	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", e)
		}
	}()

	// 获取data的数据类型
	data_v := reflect.ValueOf(datas)
	data_t := reflect.TypeOf(datas)
	data_type_name := data_t.String()
	// 查看找哪个文件
	if m, _ := regexp.MatchString(`^map\[string\]`, data_type_name); m == true {
		f_data_name = f_data_name + "_stringmap"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataStringMap{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "map[string]string":
			data.String[name] = data_v.Interface().(map[string]string)
		case "map[string]bool":
			data.Bool[name] = data_v.Interface().(map[string]bool)
		case "map[string]uint8":
			data.Uint8[name] = data_v.Interface().(map[string]uint8)
		case "map[string]uint":
			data.Uint[name] = data_v.Interface().(map[string]uint)
		case "map[string]uint64":
			data.Uint64[name] = data_v.Interface().(map[string]uint64)
		case "map[string]int8":
			data.Int8[name] = data_v.Interface().(map[string]int8)
		case "map[string]int":
			data.Int[name] = data_v.Interface().(map[string]int)
		case "map[string]int64":
			data.Int64[name] = data_v.Interface().(map[string]int64)
		case "map[string]float32":
			data.Float32[name] = data_v.Interface().(map[string]float32)
		case "map[string]float64":
			data.Float64[name] = data_v.Interface().(map[string]float64)
		case "map[string]complex64":
			data.Complex64[name] = data_v.Interface().(map[string]complex64)
		case "map[string]complex128":
			data.Complex128[name] = data_v.Interface().(map[string]complex128)
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
		// 再编码保存进去
		data_b, err = nst.StructGobBytes(data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		err = ioutil.WriteFile(f_data_name, data_b, 0600)
		if err != nil {
			err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
			return err
		}
	} else if m, _ := regexp.MatchString(`^\[\]`, data_type_name); m == true {
		f_data_name = f_data_name + "_slice"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataSlice{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "[]string":
			data.String[name] = data_v.Interface().([]string)
		case "[]bool":
			data.Bool[name] = data_v.Interface().([]bool)
		case "[]uint8":
			data.Uint8[name] = data_v.Interface().([]uint8)
		case "[]uint":
			data.Uint[name] = data_v.Interface().([]uint)
		case "[]uint64":
			data.Uint64[name] = data_v.Interface().([]uint64)
		case "[]int8":
			data.Int8[name] = data_v.Interface().([]int8)
		case "[]int":
			data.Int[name] = data_v.Interface().([]int)
		case "[]int64":
			data.Int64[name] = data_v.Interface().([]int64)
		case "[]float32":
			data.Float32[name] = data_v.Interface().([]float32)
		case "[]float64":
			data.Float64[name] = data_v.Interface().([]float64)
		case "[]complex64":
			data.Complex64[name] = data_v.Interface().([]complex64)
		case "[]complex128":
			data.Complex128[name] = data_v.Interface().([]complex128)
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
		// 再编码保存进去
		data_b, err = nst.StructGobBytes(data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		err = ioutil.WriteFile(f_data_name, data_b, 0600)
		if err != nil {
			err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
			return err
		}
	} else {
		f_data_name = f_data_name + "_normal"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataNormal{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "time.Time":
			data.Time[name] = data_v.Interface().(time.Time)
		case "[]byte":
			data.Byte[name] = data_v.Interface().([]byte)
		case "string":
			data.String[name] = data_v.Interface().(string)
		case "bool":
			data.Bool[name] = data_v.Interface().(bool)
		case "uint8":
			data.Uint8[name] = data_v.Interface().(uint8)
		case "uint":
			data.Uint[name] = data_v.Interface().(uint)
		case "uint64":
			data.Uint64[name] = data_v.Interface().(uint64)
		case "int8":
			data.Int8[name] = data_v.Interface().(int8)
		case "int":
			data.Int[name] = data_v.Interface().(int)
		case "int64":
			data.Int64[name] = data_v.Interface().(int64)
		case "float32":
			data.Float32[name] = data_v.Interface().(float32)
		case "float64":
			data.Float64[name] = data_v.Interface().(float64)
		case "complex64":
			data.Complex64[name] = data_v.Interface().(complex64)
		case "complex128":
			data.Complex128[name] = data_v.Interface().(complex128)
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
		// 再编码保存进去
		data_b, err = nst.StructGobBytes(data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		err = ioutil.WriteFile(f_data_name, data_b, 0600)
		if err != nil {
			err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
			return err
		}
	}

	return
}

// 中间类型的获取数据
func (h *HardStore) ReadDataByMiddle(id, name string, datas interface{}) (err error) {
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	//f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name

	if pubfunc.FileExist(f_name) == false {
		err = fmt.Errorf("hardstore[HardStore]WriteDataByMiddle: Can't find the Role " + id)
		return
	}

	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", e)
		}
	}()

	// 获取data的数据类型
	data_v := reflect.Indirect(reflect.ValueOf(datas))
	data_t := data_v.Type()
	data_type_name := data_t.String()
	var find bool
	var value interface{}
	// 查看找哪个文件
	if m, _ := regexp.MatchString(`^map\[string\]`, data_type_name); m == true {
		f_data_name = f_data_name + "_stringmap"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataStringMap{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "map[string]string":
			value, find = data.String[name]
		case "map[string]bool":
			value, find = data.Bool[name]
		case "map[string]uint8":
			value, find = data.Uint8[name]
		case "map[string]uint":
			value, find = data.Uint[name]
		case "map[string]uint64":
			value, find = data.Uint64[name]
		case "map[string]int8":
			value, find = data.Int8[name]
		case "map[string]int":
			value, find = data.Int[name]
		case "map[string]int64":
			value, find = data.Int64[name]
		case "map[string]float32":
			value, find = data.Float32[name]
		case "map[string]float64":
			value, find = data.Float64[name]
		case "map[string]complex64":
			value, find = data.Complex64[name]
		case "map[string]complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
	} else if m, _ := regexp.MatchString(`^\[\]`, data_type_name); m == true {
		f_data_name = f_data_name + "_slice"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataSlice{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "[]string":
			value, find = data.String[name]
		case "[]bool":
			value, find = data.Bool[name]
		case "[]uint8":
			value, find = data.Uint8[name]
		case "[]uint":
			value, find = data.Uint[name]
		case "[]uint64":
			value, find = data.Uint64[name]
		case "[]int8":
			value, find = data.Int8[name]
		case "[]int":
			value, find = data.Int[name]
		case "[]int64":
			value, find = data.Int64[name]
		case "[]float32":
			value, find = data.Float32[name]
		case "[]float64":
			value, find = data.Float64[name]
		case "[]complex64":
			value, find = data.Complex64[name]
		case "[]complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
	} else {
		f_data_name = f_data_name + "_normal"
		// 读取那个文件并转码
		data_b, err := ioutil.ReadFile(f_data_name)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		data := RoleDataNormal{}
		err = nst.BytesGobStruct(data_b, &data)
		if err != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
			return err
		}
		switch data_type_name {
		case "time.Time":
			value, find = data.Time[name]
		case "[]byte":
			value, find = data.Byte[name]
		case "string":
			value, find = data.String[name]
		case "bool":
			value, find = data.Bool[name]
		case "uint8":
			value, find = data.Uint8[name]
		case "uint":
			value, find = data.Uint[name]
		case "uint64":
			value, find = data.Uint64[name]
		case "int8":
			value, find = data.Int8[name]
		case "int":
			value, find = data.Int[name]
		case "int64":
			value, find = data.Int64[name]
		case "float32":
			value, find = data.Float32[name]
		case "float64":
			value, find = data.Float64[name]
		case "complex64":
			value, find = data.Complex64[name]
		case "complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Unsupported data type %v", data_type_name)
			return err
		}
	}
	if find == false {
		err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: Can not find the field %v", name)
		return
	}

	value_v := reflect.ValueOf(value)
	data_v.Set(value_v)
	return
}
