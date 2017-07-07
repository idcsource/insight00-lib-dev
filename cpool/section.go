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

// Create a new Section, the notes is comment.
func NewSection(name, notes string) *Section {
	return &Section{
		key:    name,
		notes:  notes,
		config: make(map[string]*Config),
		new:    true,
		del:    false,
	}
}

// Export self to encode mode.
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
			Enum:  oneconfig.enum,
			Notes: oneconfig.notes,
		}
	}
	return
}

// Decode Section and store to self.
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

// Set a Config, the s is the key, v is the value, n is the comment. if not exist, create new, else cover the old one.
func (se *Section) SetConfig(s, v, n string) error {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		se.config[s] = &Config{
			key:   s,
			notes: n,
			value: v,
			new:   true,
			del:   false,
		}
	} else {
		cf.value = v
		if len(n) != 0 {
			cf.notes = n
		}
	}
	return nil
}

// Set a Config, the s is the key, v is the value, n is the comment. if not exist, create new, else cover the old one.
func (se *Section) SetInt64(s string, v int64, n string) error {
	var vs string
	vs = strconv.FormatInt(v, 10)
	return se.SetConfig(s, vs, n)
}

// Set a Config, the s is the key, v is the value, n is the comment. if not exist, create new, else cover the old one.
func (se *Section) SetFloat(s string, v float64, n string) error {
	var vs string
	vs = strconv.FormatFloat(v, 'f', -1, 64)
	return se.SetConfig(s, vs, n)
}

// Set a Config, the s is the key, v is the value, n is the comment. if not exist, create new, else cover the old one.
func (se *Section) SetBool(s string, v bool, n string) error {
	var vs string
	if v == true {
		vs = "true"
	} else if v == false {
		vs = "false"
	}
	return se.SetConfig(s, vs, n)
}

// Set a Config, the s is the key, v is the value, n is the comment. if not exist, create new, else cover the old one.
func (se *Section) SetEnum(s string, n string, v ...string) error {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		se.config[s] = &Config{
			key:   s,
			notes: n,
			value: v[0],
			enum:  v,
			new:   true,
			del:   false,
		}
	} else {
		cf.value = v[0]
		cf.enum = v
		if len(n) != 0 {
			cf.notes = n
		}
	}
	return nil
}

// Get a Config, if not exist return error
func (se *Section) GetConfig(s string) (string, error) {
	s = strings.TrimSpace(s)
	cf, have := se.config[s]
	if have == false {
		return "", errors.New("cpool: [Section]GetConfig: Can't find the config : " + s)
	} else {
		return cf.value, nil
	}
}

// Get a Config, if not exist return error
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

// Get a Config, if not exist return error
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

// Get a Config, if not exist return error
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

// Get a Config, if not exist return error
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
