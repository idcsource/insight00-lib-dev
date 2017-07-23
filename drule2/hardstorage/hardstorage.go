// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstorage

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/idcsource/insight00-lib/iendecode"
	"github.com/idcsource/insight00-lib/pubfunc"
	"github.com/idcsource/insight00-lib/random"
	"github.com/idcsource/insight00-lib/roles"
)

// 新建一个存储实例，只需要一个存储路径和一个路径深度
func NewHardStorage(path string, deep uint8) (hardstorage *HardStorage, err error) {
	local_path := pubfunc.LocalPath(path)
	// 没有就新建
	if pubfunc.FileExist(local_path) != true {
		err = os.MkdirAll(local_path, 0700)
		if err != nil {
			err = fmt.Errorf("hardstorage[HardStorage]NewHardStorage: %v", err)
			return
		}
	} else {
		var path_info os.FileInfo
		path_info, err = os.Stat(local_path)
		if err != nil {
			err = fmt.Errorf("hardstorage[HardStorage]NewHardStorage: The local path have not be access !")
			return
		}
		if path_info.IsDir() != true {
			err = fmt.Errorf("hardstorage[HardStorage]NewHardStorage: The local path have not a path !")
			return
		}
	}
	if deep > 2 {
		deep = 2
	}
	hardstorage = &HardStorage{
		local_path: local_path,
		path_deep:  deep,
	}
	return
}

// 角色是否存在
func (h *HardStorage) RoleExist(area, roleid string) (have bool) {
	have, _ = h.roleExist(area, roleid)
	return
}

// 角色是否存在（内部），并返回角色存储的主文件名（带路径）
func (h *HardStorage) roleExist(area, roleid string) (have bool, filename string) {
	// 检查区域是否存在
	if h.existArea(area) == false {
		have = false
		return
	}
	hashid := random.GetSha1Sum(roleid)
	path := h.findRoleFilePath(area, hashid)
	filename = path + hashid
	have = pubfunc.FileExist(filename)
	return
}

// 读取角色的中间数据格式
func (h *HardStorage) RoleReadMiddleData(area, roleid string) (rolemid roles.RoleMiddleData, exist bool, err error) {
	// 查看角色是否存在
	have, f_name := h.roleExist(area, roleid)
	if have == false {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: The role not exist: %v", roleid)
		return
	}

	f_ralation_name := f_name + HARDSTORAGE_FILE_NAME_RELATION
	f_data_name := f_name + HARDSTORAGE_FILE_NAME_DATA

	f_version_b, err := ioutil.ReadFile(f_name)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}
	f_version := roles.RoleVersion{}
	err = iendecode.BytesGobStruct(f_version_b, &f_version)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}

	f_relation_b, err := ioutil.ReadFile(f_ralation_name)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}
	f_relation := roles.RoleRelation{}
	err = iendecode.BytesGobStruct(f_relation_b, &f_relation)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}

	f_data_b, err := ioutil.ReadFile(f_data_name)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}
	f_data := roles.RoleDataPoint{}
	err = iendecode.BytesGobStruct(f_data_b, &f_data)
	if err != nil {
		exist = false
		err = fmt.Errorf("hardstorage[HardStorage]RoleReadMiddleData: %v ", err)
		return
	}

	// 合成中间数据
	rolemid = roles.RoleMiddleData{
		Version:        f_version,
		VersionChange:  false,
		Relation:       f_relation,
		RelationChange: false,
		Data:           f_data,
		DataChange:     false,
	}
	exist = true
	return
}

// 增加一个角色
func (h *HardStorage) RoleAdd(area string, role roles.Roleer) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be exist.")
		return
	}

	roleid := role.ReturnId()
	// 查看是否存在这个角色
	_, f_name := h.roleExist(area, roleid)
	//	if have_role == true {
	//		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The Role already exist.")
	//		return
	//	}
	// 转码
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: %v", err)
		return
	}
	// 去存
	err = h.storeRoleMiddle(f_name, &mid)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: %v", err)
	}
	return
}

// 增加一个角色角色的中间格式
func (h *HardStorage) RoleAddMiddleData(area string, mid roles.RoleMiddleData) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be exist.")
		return
	}

	roleid := mid.Version.Id
	// 查看是否存在这个角色
	_, f_name := h.roleExist(area, roleid)
	//	if have_role == true {
	//		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The Role already exist.")
	//		return
	//	}
	// 去存
	err = h.storeRoleMiddle(f_name, &mid)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: %v", err)
	}
	return
}

// 存储一个角色角色的中间格式（不检查是否存在）
func (h *HardStorage) RoleStoreMiddleData(area string, mid roles.RoleMiddleData) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: The area name not be exist.")
		return
	}

	roleid := mid.Version.Id
	// 获取文件名
	_, f_name := h.roleExist(area, roleid)
	// 去存
	err = h.storeRoleMiddle(f_name, &mid)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: %v", err)
	}
	return
}

func (h *HardStorage) RoleUpdate(area string, role roles.Roleer) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The area name not be exist.")
		return
	}

	roleid := role.ReturnId()
	// 查看是否存在这个角色
	have_role, f_name := h.roleExist(area, roleid)
	if have_role == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The Role not exist.")
		return
	}

	// 转码
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleStoreMiddleData: %v", err)
		return
	}

	// 去存
	err = h.storeRoleMiddle(f_name, &mid)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: %v", err)
	}
	return
}

// 更新角色的中间格式（不存在则错误）
func (h *HardStorage) RoleUpdateMiddleData(area string, mid roles.RoleMiddleData) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The area name not be exist.")
		return
	}

	roleid := mid.Version.Id
	// 查看是否存在这个角色
	have_role, f_name := h.roleExist(area, roleid)
	if have_role == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: The Role not exist.")
		return
	}
	// 去存
	err = h.storeRoleMiddle(f_name, &mid)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]RoleUpdateMiddleData: %v", err)
	}
	return
}

// 存储方法
func (h *HardStorage) storeRoleMiddle(f_name string, mid *roles.RoleMiddleData) (err error) {
	f_ralation_name := f_name + HARDSTORAGE_FILE_NAME_RELATION
	f_data_name := f_name + HARDSTORAGE_FILE_NAME_DATA

	if mid.VersionChange == true {
		// 保存Version
		version_b, err := iendecode.StructGobBytes(mid.Version)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(f_name, version_b, 0600)
		if err != nil {
			return err
		}
	}

	if mid.DataChange == true {
		// 存数据
		data_b, err := iendecode.StructGobBytes(mid.Data)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(f_data_name, data_b, 0600)
		if err != nil {
			return err
		}
	}

	if mid.RelationChange == true {
		// 保存关系
		relation_b, err := iendecode.StructGobBytes(mid.Relation)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(f_ralation_name, relation_b, 0600)
		if err != nil {
			return err
		}
	}
	return
}

// 删除角色
func (h *HardStorage) RoleDelete(area, roleid string) (err error) {
	// 查看是否存在这个角色
	have_role, f_name := h.roleExist(area, roleid)
	if have_role == false {
		err = fmt.Errorf("hardstorage[HardStorage]RoleDelete: The Role not exist.")
		return
	}

	f_ralation_name := f_name + HARDSTORAGE_FILE_NAME_RELATION
	f_data_name := f_name + HARDSTORAGE_FILE_NAME_DATA

	if pubfunc.FileExist(f_name) == true {
		os.Remove(f_name)
	}
	if pubfunc.FileExist(f_ralation_name) == true {
		os.Remove(f_ralation_name)
	}
	if pubfunc.FileExist(f_data_name) == true {
		os.Remove(f_data_name)
	}

	return
}

// 区域改名
func (h *HardStorage) AreaReName(oldname, newname string) (err error) {
	// 查看老的是否存在
	if h.existArea(oldname) == false {
		err = fmt.Errorf("harstorage[HardStorage]AreaReName: The area not exist.")
		return
	}
	// 查看新的是否合法
	if CheckAreaName(newname) == false {
		err = fmt.Errorf("harstorage[HardStorage]AreaReName: The new area name not allow.")
		return
	}
	// 查看新的是否存在
	if h.existArea(newname) == true {
		err = fmt.Errorf("harstorage[HardStorage]AreaReName: The new area name is be used.")
		return
	}
	// 换名
	oldpath := h.local_path + oldname
	newpath := h.local_path + newname
	err = os.Rename(oldpath, newpath)
	if err != nil {
		err = fmt.Errorf("harstorage[HardStorage]AreaReName: %v", err)
	}

	return
}

// 删除区域
func (h *HardStorage) AreaDelete(area string) (err error) {
	if h.existArea(area) == false {
		return
	}
	path := h.local_path + area
	err = os.RemoveAll(path)
	if err != nil {
		err = fmt.Errorf("harstorage[HardStorage]AreaDelete: %v", err)
	}
	return
}

// 初始化区域
func (h *HardStorage) AreaInit(path string) (err error) {
	// 查看是否存在
	if h.existArea(path) {
		err = fmt.Errorf("hardstorage[HardStorage]AreaInit: The area already exist.")
		return
	}
	// 创建
	err = h.createArea(path)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStorage]AreaInit: %v", err)
	}
	return
}

// 返回区域列表
func (h *HardStorage) AreaList() (list []string, err error) {
	dir, err := ioutil.ReadDir(h.local_path)
	if err != nil {
		err = fmt.Errorf("hardstorage[HardStroage]AreaList: %v", err)
		return
	}
	list = make([]string, 0)
	for _, path := range dir {
		if path.IsDir() {
			list = append(list, path.Name())
		}
	}
	return
}

// 创建区域路径
func (h *HardStorage) createArea(path string) (err error) {
	// 检查区域名称是否合法
	allow := CheckAreaName(path)
	if allow == false {
		err = fmt.Errorf("The area name not be allow.")
		return
	}

	l_path_name := []string{"a", "b", "c", "d", "e", "f", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	l_path := make([]string, 0)

	var path_deep, i uint8
	path_deep = h.path_deep
	local_path := h.local_path + path + "/"
	for i = 0; i < path_deep; i++ {
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
			err = os.MkdirAll(v, 0700)
			if err != nil {
				return
			}
		}
	}
	return
}

// 区域是否存在
func (h *HardStorage) AreaExist(area string) (have bool) {
	return h.existArea(area)
}

// 查看区域是否存在
func (h *HardStorage) existArea(path string) (have bool) {
	allow := CheckAreaName(path)
	if allow == false {
		return false
	}
	have = pubfunc.FileExist(h.local_path + path)
	return
}

// 根据角色的Id找到存放路径
func (h *HardStorage) findRoleFilePath(area, hashid string) (path string) {
	path = h.local_path + area + "/"
	for i := 0; i < int(h.path_deep); i++ {
		path = path + string(hashid[i]) + "/"
	}
	return path
}
