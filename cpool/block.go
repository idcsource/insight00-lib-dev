// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license


package cpool

import (
	"strings"
	"errors"
	"strconv"
)

// 新建一个块（Block）配置
func NewBlock (name, notes string) *Block {
	return &Block{
		key : name,
		notes : notes,
		section: make(map[string]*Section),
		new : true,
		del : false,
	};
}

// 编码自身
func (b *Block) EncodeBlock () (encode BlockEncode) {
	encode = BlockEncode{
		Key: b.key,
		Notes : b.notes,
		Sections: make(map[string]SectionEncode),
	};
	for sectionkey, onesection := range b.section {
		encode.Sections[sectionkey] = SectionEncode{
			Key: onesection.key,
			Notes: onesection.notes,
			Configs: make(map[string]ConfigEncode),
		};
		for configkey, oneconfig := range onesection.config {
			encode.Sections[sectionkey].Configs[configkey] = ConfigEncode{
				Key : oneconfig.key,
				Value : oneconfig.value,
				Notes : oneconfig.notes,
			}
		}
	}
	return;
}

// 解码block并装入自己
func (b *Block) DecodeBlock (block BlockEncode) {
	b.key = block.Key;
	b.notes = block.Notes;
	for key, onesection := range block.Sections {
		b.section[key] = &Section{
			key : onesection.Key,
			notes : onesection.Notes,
			config : make(map[string]*Config),
			new : true,
			del : false,
		};
		b.section[key].DecodeSection(onesection);
	}
}

// 解码一个section并装入相应的位置，如果section已经存在将会被覆盖
func (b *Block) DecodeSection (se SectionEncode) {
	sname := se.Key;
	_, find := b.section[sname];
	if find == false {
		b.section[sname] = &Section{
			key : se.Key,
			notes : se.Notes,
			new : true,
			del : false,
		}
	}
	b.section[sname].DecodeSection(se);
	if find == true {
		b.section[sname].new = false;
	}
}

// 将某个片导出为编码模式，如果片不存在，则返回错误
func (b *Block) EncodeSection (s string) (encode SectionEncode, err error) {
	encode = SectionEncode{
		Configs: make(map[string]ConfigEncode),
	}
	s = strings.TrimSpace(s);
	onesection, ok := b.section[s];
	if ok == false{
		err = errors.New("cpool: [Block]EncodeSection: Can't find the Section : " + s);
		return;
	}
	encode.Key = onesection.key;
	encode.Notes = onesection.notes;
	for configkey, oneconfig := range onesection.config {
		encode.Configs[configkey] = ConfigEncode{
			Key : oneconfig.key,
			Value : oneconfig.value,
			Notes : oneconfig.notes,
		}
	}
	return;
}

// 获取一个片(Section)内的内容，s为片的名称
func (b *Block) GetSection (s string)(*Section, error) {
	s = strings.TrimSpace(s);
	rsAs, ok := b.section[s];
	if ok == false{
		return nil, errors.New("cpool: [Block]GetSection: Can't find the config : " + s);
	}
	return rsAs, nil;
}

// 将自己创建的片（Section）注册进块（Block）中，如果块中已经有同名的则返回错误
func (b *Block) RegSection (s *Section) error {
	skey := s.key;
	_, find := b.section[skey];
	if find == true {
		return errors.New("cpool: [Block]RegSection: The Section " + skey + " is already exist.");
	}
	b.section[skey] = s;
	return nil;
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
//
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (b *Block) SetConfig (s, v, n string) error {
	s = strings.TrimSpace(s);
	csA := strings.Split(s,".");
	if len(csA) != 2 {
		return errors.New("cpool: [Block]SetConfig: Request configuration node error : " + s);
	}
	csA1 := strings.TrimSpace(csA[0]);
	csA2 := strings.TrimSpace(csA[1]);
	rsA, ok := b.section[csA1];
	if ok == false{
		b.section[csA1] = &Section{
			key : csA1,
			config : make(map[string]*Config),
			new : true,
			del : false,
		};
		rsA = b.section[csA1];
	}
	rs, ok2 := rsA.config[csA2];
	if ok2 == false{
		rsA.config[csA2] = &Config{
			key : csA2,
			value : v,
			notes : "#" + n,
			new : true,
			del : false,
		};
	}else{
		rs.value = v;
		if len(n) != 0 {
			rs.notes = "#" + n;
		}
	}
	return nil;
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (b *Block) SetInt64 (s string, v int64, n string) error {
	var vs string;
	vs = strconv.FormatInt(v,10);
	return b.SetConfig(s, vs, n);
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (b *Block) SetFloat (s string, v float64, n string) error {
	var vs string;
	vs = strconv.FormatFloat(v,'f',-1,64);
	return b.SetConfig(s, vs, n);
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (b *Block) SetBool (s string, v bool, n string) error {
	var vs string;
	if v == true {
		vs = "true";
	}else if v == false {
		vs = "false";
	}
	return b.SetConfig(s, vs, n);
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
// 注意这里的注释为第二个参数
func (b *Block) SetEnum (s string, n string, v ...string) error {
	var vs string;
	vs = strings.Join(v,",");
	return b.SetConfig(s, vs, n);
}

// 获取一个配置，没有找到将返回错误，s格式为section.keyname
func (b *Block) GetConfig (s string) (string, error) {
	s = strings.TrimSpace(s);
	csA := strings.Split(s,".");
	if len(csA) != 2 {
		return "", errors.New("cpool: [Block]GetConfig: Request configuration node error : " + s);
	}
	csA1 := strings.TrimSpace(csA[0]);
	csA2 := strings.TrimSpace(csA[1]);
	rsA, ok := b.section[csA1];
	if ok == false{
		return "", errors.New("cpool: [Block]GetConfig: Can't find the config : " + s);
	}
	rs, ok2 := rsA.config[csA2];
	if ok2 == false{
		return "", errors.New("cpool: [Block]GetConfig: Can't find the config : " + s);
	}
	return rs.value, nil;
}

// 获取一个配置，并转换为字符串切片，转换失败或没有找到将返回错误，s格式为section.keyname
func (b *Block) GetEnum (s string) ([]string, error) {
	config, err := b.GetConfig(s);
	if err != nil {
		return nil, err;
	}
	configa := strings.Split(config,",");
	returna := make([]string,0);
	for _, v := range configa {
		v = strings.TrimSpace(v);
		if len(v) > 0 {
			returna = append(returna, v);
		}
	}
	return returna, nil;
}

// 获取一个配置，并转换为64为整形，转换失败或没有找到将返回错误，s格式为section.keyname
func (b *Block) TranInt64 (s string) (int64, error) {
	cf, err := b.GetConfig(s);
	if err != nil {
		return 0, errors.New("cpool: [Block]TranInt64: Request configuration node error : " + s);
	}
	i, err2 := strconv.ParseInt(cf,10,64);
	if err2 != nil {
		return 0, err2;
	}
	return i, nil;
}

// 获取一个配置，并转换为64为浮点，转换失败或没有找到将返回错误，s格式为section.keyname
func (b *Block) TranFloat (s string) (float64, error) {
	cf, err := b.GetConfig(s);
	if err != nil {
		return 0, errors.New("cpool: [Block]TranFloat: Request configuration node error : " + s);
	}
	i, err2 := strconv.ParseFloat(cf,64);
	if err2 != nil {
		return 0, err2;
	}
	return i, nil;
}

// 获取一个配置，并转换为布尔值，转换失败或没有找到将返回错误，s格式为section.keyname
func(b *Block) TranBool (s string) (bool, error) {
	cf, err := b.GetConfig(s);
	if err != nil {
		return false, errors.New("cpool: [Block]TranBool: Request configuration node error : " + s);
	}
	cf = strings.ToLower(cf);
	if cf == "true" || cf == "yes" || cf == "t" || cf == "y" {
		return true, nil;
	}else if cf == "false" || cf == "no" || cf == "f" || cf == "n" {
		return false, nil;
	}else{
		return false, errors.New("cpool: [Block]TranBool: Not bool : " + s);
	}
}
