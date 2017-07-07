// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package cpool

import (
	"errors"
	"strconv"
	"strings"
)

// Create a new Block
func NewBlock(name, notes string) *Block {
	return &Block{
		key:     name,
		notes:   notes,
		section: make(map[string]*Section),
		new:     true,
		del:     false,
	}
}

// Encode self
func (b *Block) EncodeBlock() (encode BlockEncode) {
	encode = BlockEncode{
		Key:      b.key,
		Notes:    b.notes,
		Sections: make(map[string]SectionEncode),
	}
	for sectionkey, onesection := range b.section {
		encode.Sections[sectionkey] = SectionEncode{
			Key:     onesection.key,
			Notes:   onesection.notes,
			Configs: make(map[string]ConfigEncode),
		}
		for configkey, oneconfig := range onesection.config {
			encode.Sections[sectionkey].Configs[configkey] = ConfigEncode{
				Key:   oneconfig.key,
				Value: oneconfig.value,
				Enum:  oneconfig.enum,
				Notes: oneconfig.notes,
			}
		}
	}
	return
}

// decode Block and store to self
func (b *Block) DecodeBlock(block BlockEncode) {
	b.key = block.Key
	b.notes = block.Notes
	for key, onesection := range block.Sections {
		b.section[key] = &Section{
			key:    onesection.Key,
			notes:  onesection.Notes,
			config: make(map[string]*Config),
			new:    true,
			del:    false,
		}
		b.section[key].DecodeSection(onesection)
	}
}

// decode a Section and store, if the Section name is exist, cover the old one.
func (b *Block) DecodeSection(se SectionEncode) {
	sname := se.Key
	_, find := b.section[sname]
	if find == false {
		b.section[sname] = &Section{
			key:   se.Key,
			notes: se.Notes,
			new:   true,
			del:   false,
		}
	}
	b.section[sname].DecodeSection(se)
	if find == true {
		b.section[sname].new = false
	}
}

// export a Section to encode mode
func (b *Block) EncodeSection(s string) (encode SectionEncode, err error) {
	encode = SectionEncode{
		Configs: make(map[string]ConfigEncode),
	}
	s = strings.TrimSpace(s)
	onesection, ok := b.section[s]
	if ok == false {
		err = errors.New("cpool: [Block]EncodeSection: Can't find the Section : " + s)
		return
	}
	encode.Key = onesection.key
	encode.Notes = onesection.notes
	for configkey, oneconfig := range onesection.config {
		encode.Configs[configkey] = ConfigEncode{
			Key:   oneconfig.key,
			Value: oneconfig.value,
			Enum:  oneconfig.enum,
			Notes: oneconfig.notes,
		}
	}
	return
}

// Get a Section
func (b *Block) GetSection(s string) (*Section, error) {
	s = strings.TrimSpace(s)
	rsAs, ok := b.section[s]
	if ok == false {
		return nil, errors.New("cpool: [Block]GetSection: Can't find the config : " + s)
	}
	return rsAs, nil
}

// register a Section,  if the Section name is exist, return err.
func (b *Block) RegSection(s *Section) error {
	skey := s.key
	_, find := b.section[skey]
	if find == true {
		return errors.New("cpool: [Block]RegSection: The Section " + skey + " is already exist.")
	}
	b.section[skey] = s
	return nil
}

// Set one config， the s format is ：block|section.configkey， v is the value， n is the comment.
//
// if the Config exist, cover the old one.
func (b *Block) SetConfig(s, v, n string) error {
	s = strings.TrimSpace(s)
	csA := strings.Split(s, ".")
	if len(csA) != 2 {
		return errors.New("cpool: [Block]SetConfig: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0])
	csA2 := strings.TrimSpace(csA[1])
	rsA, ok := b.section[csA1]
	if ok == false {
		b.section[csA1] = &Section{
			key:    csA1,
			config: make(map[string]*Config),
			new:    true,
			del:    false,
		}
		rsA = b.section[csA1]
	}
	rs, ok2 := rsA.config[csA2]
	if ok2 == false {
		rsA.config[csA2] = &Config{
			key:   csA2,
			value: v,
			notes: n,
			new:   true,
			del:   false,
		}
	} else {
		rs.value = v
		if len(n) != 0 {
			rs.notes = n
		}
	}
	return nil
}

// Set one config， the s format is ：block|section.configkey， v is the value， n is the comment.
//
// if the Config exist, cover the old one.
func (b *Block) SetInt64(s string, v int64, n string) error {
	var vs string
	vs = strconv.FormatInt(v, 10)
	return b.SetConfig(s, vs, n)
}

// Set one config， the s format is ：block|section.configkey， v is the value， n is the comment.
//
// if the Config exist, cover the old one.
func (b *Block) SetFloat(s string, v float64, n string) error {
	var vs string
	vs = strconv.FormatFloat(v, 'f', -1, 64)
	return b.SetConfig(s, vs, n)
}

// Set one config， the s format is ：block|section.configkey， v is the value， n is the comment.
//
// if the Config exist, cover the old one.
func (b *Block) SetBool(s string, v bool, n string) error {
	var vs string
	if v == true {
		vs = "true"
	} else if v == false {
		vs = "false"
	}
	return b.SetConfig(s, vs, n)
}

// Set one config， the s format is ：block|section.configkey， v is the value， n is the comment.
//
// if the Config exist, cover the old one.
func (b *Block) SetEnum(s string, n string, v ...string) error {
	s = strings.TrimSpace(s)
	csA := strings.Split(s, ".")
	if len(csA) != 2 {
		return errors.New("cpool: [Block]SetEnum: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0])
	csA2 := strings.TrimSpace(csA[1])
	rsA, ok := b.section[csA1]
	if ok == false {
		b.section[csA1] = &Section{
			key:    csA1,
			config: make(map[string]*Config),
			new:    true,
			del:    false,
		}
		rsA = b.section[csA1]
	}
	rs, ok2 := rsA.config[csA2]
	if ok2 == false {
		rsA.config[csA2] = &Config{
			key:   csA2,
			value: v[0],
			enum:  v,
			notes: n,
			new:   true,
			del:   false,
		}
	} else {
		rs.enum = v
		rs.value = v[0]
		if len(n) != 0 {
			rs.notes = n
		}
	}
	return nil
}

// get a Config, if not exist, return error, s format is section.keyname
func (b *Block) GetConfig(s string) (string, error) {
	s = strings.TrimSpace(s)
	csA := strings.Split(s, ".")
	if len(csA) != 2 {
		return "", errors.New("cpool: [Block]GetConfig: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0])
	csA2 := strings.TrimSpace(csA[1])
	rsA, ok := b.section[csA1]
	if ok == false {
		return "", errors.New("cpool: [Block]GetConfig: Can't find the config : " + s)
	}
	rs, ok2 := rsA.config[csA2]
	if ok2 == false {
		return "", errors.New("cpool: [Block]GetConfig: Can't find the config : " + s)
	}
	return rs.value, nil
}

// get a enum config, if not exist, return error. the s format is section.keyname
func (b *Block) GetEnum(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	csA := strings.Split(s, ".")
	if len(csA) != 2 {
		return nil, errors.New("cpool: [Block]GetEnum: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0])
	csA2 := strings.TrimSpace(csA[1])
	rsA, ok := b.section[csA1]
	if ok == false {
		return nil, errors.New("cpool: [Block]GetEnum: Can't find the config : " + s)
	}
	rs, ok2 := rsA.config[csA2]
	if ok2 == false {
		return nil, errors.New("cpool: [Block]GetEnum: Can't find the config : " + s)
	}
	return rs.enum, nil
	//	config, err := b.GetConfig(s);
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

// get a Config, and transform to int64, if not exist return error. the s format is section.keyname
func (b *Block) TranInt64(s string) (int64, error) {
	cf, err := b.GetConfig(s)
	if err != nil {
		return 0, errors.New("cpool: [Block]TranInt64: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseInt(cf, 10, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil
}

// get a Config, and transform to float64, if not exist return error. the s format is section.keyname
func (b *Block) TranFloat(s string) (float64, error) {
	cf, err := b.GetConfig(s)
	if err != nil {
		return 0, errors.New("cpool: [Block]TranFloat: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseFloat(cf, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil
}

// get a Config, and transform to bool, if not exist return error. the s format is section.keyname
func (b *Block) TranBool(s string) (bool, error) {
	cf, err := b.GetConfig(s)
	if err != nil {
		return false, errors.New("cpool: [Block]TranBool: Request configuration node error : " + s)
	}
	cf = strings.ToLower(cf)
	if cf == "true" || cf == "yes" || cf == "t" || cf == "y" {
		return true, nil
	} else if cf == "false" || cf == "no" || cf == "f" || cf == "n" {
		return false, nil
	} else {
		return false, errors.New("cpool: [Block]TranBool: Not bool : " + s)
	}
}
