// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package hardstorage

import (
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/idcsource/insight00-lib/drule2/types"
	"github.com/idcsource/insight00-lib/pubfunc"
	"github.com/idcsource/insight00-lib/random"
	"github.com/idcsource/insight00-lib/spots"

	"github.com/cznic/zappy"
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
func (h *HardStorage) SpotExist(area, spotid string) (have bool) {
	have, _ = h.spotExist(area, spotid)
	return
}

// 角色是否存在（内部），并返回角色存储的主文件名（带路径）
func (h *HardStorage) spotExist(area, spotid string) (have bool, filename string) {
	// 检查区域是否存在
	if h.existArea(area) == false {
		have = false
		return
	}
	hashid := random.GetSha1Sum(spotid)
	path := h.findSpotFilePath(area, hashid)
	filename = path + hashid
	have = pubfunc.FileExist(filename)
	return
}

func (h *HardStorage) SpotStore(area string, spot *spots.Spots) (err error) {
	// 检查区域是否合法
	allow := CheckAreaName(area)
	if allow == false {
		err = fmt.Errorf("hardstorage[HardStorage]SpotStore: The area name not be allow.")
		return
	}
	// 检查区域是否存在
	have_a := h.existArea(area)
	if have_a == false {
		err = fmt.Errorf("hardstorage[HardStorage]SpotStore: The area name not be exist.")
		return
	}
	spotid := spot.GetId()
	_, f_name := h.spotExist(area, spotid)
	// to bytes
	spot_b, err := spot.MarshalBinary()
	if err != nil {
		return
	}
	// to zip
	spot_b_zip, err := zappy.Encode(nil, spot_b)
	if err != nil {
		return
	}
	// to write
	err = ioutil.WriteFile(f_name, spot_b_zip, 0600)
	return
}

func (h *HardStorage) SpotRead(area, spotid string) (spot *spots.Spots, err error) {
	have, f_name := h.spotExist(area, spotid)
	if have == false {
		err = fmt.Errorf("hardstorage[HardStorage]SpotRead: The Spot not exist: %v", spotid)
		return
	}
	// to read
	spot_b_zip, err := ioutil.ReadFile(f_name)
	if err != nil {
		return
	}
	// to unzip
	spot_b, err := zappy.Decode(nil, spot_b_zip)
	if err != nil {
		return
	}
	spot = spots.NewEmptySpot()
	err = spot.UnmarshalBinary(spot_b)
	return
}

func (h *HardStorage) SpotReadWithBody(area, spotid string, body spots.DataBodyer) (spot *spots.Spots, err error) {
	spot, err = h.SpotRead(area, spotid)
	if err != nil {
		return
	}
	err = spot.BtoDataBody(body)
	return
}

func (h *HardStorage) SpotDelete(area, spotid string) (err error) {
	have, f_name := h.spotExist(area, spotid)
	if have == false {
		err = fmt.Errorf("hardstorage[HardStorage]SpotDelete: The Spot not exist.")
		return
	}
	if pubfunc.FileExist(f_name) == true {
		os.Remove(f_name)
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
func (h *HardStorage) findSpotFilePath(area, hashid string) (path string) {
	path = h.local_path + area + "/"
	for i := 0; i < int(h.path_deep); i++ {
		path = path + string(hashid[i]) + "/"
	}
	return path
}
