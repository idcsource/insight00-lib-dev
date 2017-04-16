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

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/iendecode"
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
func (h *HardStore) ReadMiddleData(id string) (mid roles.RoleMiddleData, err error) {
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
	f_version := roles.RoleVersion{}
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
	f_relation := roles.RoleRelation{}
	err = nst.BytesGobStruct(f_relation_b, &f_relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_normal_b, err := ioutil.ReadFile(f_data_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_normal := roles.RoleDataPoint{}
	err = nst.BytesGobStruct(f_data_normal_b, &f_data_normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	// 合成中间数据
	mid = roles.RoleMiddleData{
		Version:  f_version,
		Relation: f_relation,
		Data:     f_data_normal,
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
	f_version := roles.RoleVersion{}
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
	f_relation := roles.RoleRelation{}
	err = nst.BytesGobStruct(f_relation_b, &f_relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	f_data_normal_b, err := ioutil.ReadFile(f_data_name)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}
	f_data_normal := roles.RoleDataPoint{}
	err = nst.BytesGobStruct(f_data_normal_b, &f_data_normal)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
		return
	}

	// 合成中间数据
	mid := roles.RoleMiddleData{
		Version:  f_version,
		Relation: f_relation,
		Data:     f_data_normal,
	}
	// 解码
	err = roles.DecodeMiddleToRole(mid, role)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]ReadRoleByMiddle: %v " + id)
	}
	return
}

// 解码一个角色，将二进制的角色存储进行解码
func (h *HardStore) DecodeRole(roleb, relab, verb []byte) (role roles.Roleer, err error) {
	return DecodeRole(roleb, relab, verb)
}

func (h *HardStore) DecodeMiddleToRole(mid roles.RoleMiddleData, role roles.Roleer) (err error) {
	return roles.DecodeMiddleToRole(mid, role)
}

// 存储角色，直接从[]byte结构
func (h *HardStore) StoreRoleByMiddleByte(b []byte) (err error) {
	mid := roles.RoleMiddleData{}
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
	data_b, err := nst.StructGobBytes(mid.Data)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddleByte: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name, data_b, 0600)
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

func (h *HardStore) StoreRoleFromMiddle(mid *roles.RoleMiddleData) (err error) {
	id := mid.Version.Id
	hashid := random.GetSha1Sum(id)
	path := h.findRoleFilePath(hashid)
	f_name := path + hashid
	f_ralation_name := f_name + h.relation_name
	f_data_name := f_name + h.data_name
	// 存数据
	data_b, err := nst.StructGobBytes(mid.Data)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name, data_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
		return
	}

	// 保存关系和Version
	version_b, err := nst.StructGobBytes(mid.Version)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
		return
	}
	relation_b, err := nst.StructGobBytes(mid.Relation)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_ralation_name, relation_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_name, version_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleFromMiddle: %v", err)
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
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}

	// 存数据
	data_b, err := nst.StructGobBytes(mid.Data)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]StoreRoleByMiddle: %v", err)
		return
	}
	err = ioutil.WriteFile(f_data_name, data_b, 0600)
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
	return nil
}

// 编码角色，将角色编码为两个部分的[]byte，一个是角色本身的数据roleb，一个是角色的关系relab
func (h *HardStore) EncodeRole(role roles.Roleer) (roleb, relab, verb []byte, err error) {
	return EncodeRole(role)
}

func (h *HardStore) EncodeRoleToMiddle(role roles.Roleer) (mid roles.RoleMiddleData, err error) {
	return roles.EncodeRoleToMiddle(role)
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
	// 读取那个文件并转码
	data_b, err := ioutil.ReadFile(f_data_name)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
		return err
	}
	data := roles.RoleDataPoint{}
	err = nst.BytesGobStruct(data_b, &data)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
		return err
	}
	datas_b, err := iendecode.StructGobBytes(datas)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
		return err
	}
	data.Point[name] = datas_b
	// 再编码保存进去
	data_b, err = nst.StructGobBytes(data)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]WriteDataByMiddle: %v", err)
		return err
	}
	err = ioutil.WriteFile(f_data_name, data_b, 0600)
	if err != nil {
		err = fmt.Errorf("hardstore[HardStore]WriteDataByMiddle: %v", err)
		return err
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
		err = fmt.Errorf("hardstore[HardStore]ReadDataByMiddle: Can't find the Role " + id)
		return
	}
	data_b, err := ioutil.ReadFile(f_data_name)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]ReadDataByMiddle: %v", err)
		return err
	}
	data := roles.RoleDataPoint{}
	err = nst.BytesGobStruct(data_b, &data)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]ReadDataByMiddle: %v", err)
		return err
	}
	_, find := data.Point[name]
	if find == false {
		err = fmt.Errorf("hardstore[RoleMiddleData]ReadDataByMiddle: Can not find the field %v", name)
		return
	}
	err = iendecode.BytesGobStruct(data.Point[name], datas)
	if err != nil {
		err = fmt.Errorf("hardstore[RoleMiddleData]ReadDataByMiddle: %v", err)
	}
	return
}
