// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package cpool

import (
	"errors"
	"strconv"
	"strings"
)

// 新建一个片配置
func NewSection(name, notes string) *Section {
	return &Section{
		key:    name,
		notes:  notes,
		config: make(map[string]*Config),
		new:    true,
		del:    false,
	}
}

// 将自身导出为编码模式
func (se *Section) EncodeSection() (encode SectionEncode) {
	encode = SectionEncode{
		Key:     se.key,
		Notes:   se.notes,
		Configs: make(map[string]ConfigEncode),
	}
	for configkey, oneconfig := range se.config {
		encode.Configs[configkey] = ConfigEncode{
			Key:   oneconfig.key,
			Value: oneconfig.value,
			Notes: oneconfig.notes,
		}
	}
	return
}

// 解码section并装入自己
func (se *Section) DecodeSection(s SectionEncode) {
	se.key = s.Key
	se.notes = s.Notes
	for key, oneconfig := range s.Configs {
		_, find := se.config[key]
		se.config[key] = &Config{
			key:   oneconfig.Key,
			value: oneconfig.Value,
			notes: oneconfig.Notes,
			new:   true,
			del:   false,
		}
		if find == true {
			se.config[key].new = false
		}
	}
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (se *Section) SetConfig(s, v, n string) error {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		se.config[s] = &Config{
			key:   s,
			notes: "#" + n,
			value: v,
			new:   true,
			del:   false,
		}
	} else {
		cf.value = v
		if len(n) != 0 {
			cf.notes = "#" + n
		}
	}
	return nil
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (se *Section) SetInt64(s string, v int64, n string) error {
	var vs string
	vs = strconv.FormatInt(v, 10)
	return se.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (se *Section) SetFloat(s string, v float64, n string) error {
	var vs string
	vs = strconv.FormatFloat(v, 'f', -1, 64)
	return se.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (se *Section) SetBool(s string, v bool, n string) error {
	var vs string
	if v == true {
		vs = "true"
	} else if v == false {
		vs = "false"
	}
	return se.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
// 注意这里的注释为第二个参数
func (se *Section) SetEnum(s string, n string, v ...string) error {
	var vs string
	vs = strings.Join(v, ",")
	return se.SetConfig(s, vs, n)
}

// 获取一个配置，没有找到将返回错误，s为配置的键名
func (se *Section) GetConfig(s string) (string, error) {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		return "", errors.New("cpool: [Section]GetConfig: Can't find the config : " + s)
	} else {
		return cf.value, nil
	}
}

// 获取一个配置，并转换成字符串切片，转换失败或没有找到将返回错误，s为配置的键名
func (se *Section) GetEnum(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		return nil, errors.New("cpool: [Section]GetEnum: Can't find the config : " + s)
	} else {
		return cf.enum, nil
	}
	//	config, err := se.GetConfig(s);
	//	if err != nil {
	//		return nil, err;
	//	}
	//	configa := strings.Split(config,",");
	//	returna := make([]string,0);
	//	for _, v := range configa {
	//		v = strings.TrimSpace(v);
	//		if len(v) > 0 {
	//			returna = append(returna, v);
	//		}
	//	}
	//	return returna, nil;
}

// 获取一个配置，并转换为64为整形，转换失败或没有找到将返回错误，s为配置的键名
func (se *Section) TranInt64(s string) (int64, error) {
	cf, err := se.GetConfig(s)
	if err != nil {
		return 0, errors.New("cpool: [Section]TranInt64: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseInt(cf, 10, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil
}

// 获取一个配置，并转换为64为浮点，转换失败或没有找到将返回错误，s为配置的键名
func (se *Section) TranFloat(s string) (float64, error) {
	cf, err := se.GetConfig(s)
	if err != nil {
		return 0, errors.New("cpool: [Section]TranFloat: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseFloat(cf, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil
}

// 获取一个配置，并转换为布尔值，转换失败或没有找到将返回错误，s为配置的键名
func (se *Section) TranBool(s string) (bool, error) {
	cf, err := se.GetConfig(s)
	if err != nil {
		return false, errors.New("cpool: [Section]TranBool: Request configuration node error : " + s)
	}
	cf = strings.ToLower(cf)
	if cf == "true" || cf == "yes" || cf == "t" || cf == "y" {
		return true, nil
	} else if cf == "false" || cf == "no" || cf == "f" || cf == "n" {
		return false, nil
	} else {
		return false, errors.New("cpool: [Section]TranBool: Not bool : " + s)
	}
}
