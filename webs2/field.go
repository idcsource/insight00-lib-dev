// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

const (
	// 表单输入项内容
	FIELD_FORM_TEXT         = iota // 文本：test
	FIELD_FORM_TEXTAREA            // 文本域：testara
	FIELD_FORM_MARKDOWN            // Markdown文本：markdown
	FIELD_FORM_EDITOR              // 编辑器：editor
	FIELD_FORM_MARK                // 也就是A-Za-z0-9_-：mark
	FIELD_FORM_PATH                // 路径：path
	FIELD_FORM_LINK                // 链接：也就是http://或https://或ftp://开头：link
	FIELD_FORM_EMAIL               // 邮箱：email
	FIELD_FORM_FILE                // 文件：file
	FIELD_FORM_TIME                // 时间：time
	FIELD_FORM_DATE                // 日期：date
	FIELD_FORM_INT                 // 数字:int
	FIELD_FORM_FLOAT               // 浮点:float
	FIELD_FORM_MONEY               // 钱:money
	FIELD_FORM_PASSWORD            // 密码:password
	FIELD_FORM_PASSWORD_TWO        // 密码两遍输入:passwordt
	FIELD_FORM_ENUM_STRING         // 字符串类枚举:enums
	FIELD_FORM_ENUM_INT            // 数字类枚举:enumi
	FIELD_FORM_ENUM_FLOAT          // 浮点型枚举:enumf
	FIELD_FORM_OTHER               // 未知类型
)

type FieldConfig struct {
	DatabaseField string    //数据库字段
	UseIt         bool      //是否启用
	DisName       string    //显示名称
	Type          uint8     //类型
	CanNull       bool      //是否可以为空
	Info          string    //说明信息
	Min           int64     //最小长度或最小数字
	Max           int64     //最大长度或最大数字
	EnumString    []string  //字符串枚举
	EnumInt       []int64   //数字型枚举
	EnumFloat     []float64 //浮点型枚举
	Other         uint8     //其他状态
}

type AllFieldsConfig map[string]*FieldConfig

func GetFieldConfig(config *cpool.Section) (afc AllFieldsConfig) {
	afc = make(map[string]*FieldConfig)
	c_encode := config.EncodeSection()
	for k, onefield := range c_encode.Configs {
		ofc := &FieldConfig{DatabaseField: k}
		getOneFieldConfig(ofc, onefield.Value)
		afc[k] = ofc
	}
	return afc
}

func getOneFieldConfig(field *FieldConfig, config string) {
	cfA := strings.Split(config, ",")
	for i, ocf := range cfA {
		switch i {
		case 0:
			//是否使用
			if ocf == "true" {
				field.UseIt = true
			} else {
				field.UseIt = false
				continue
			}
		case 1:
			//显示名
			field.DisName = ocf
		case 2:
			//类型
			switch ocf {
			case "text":
				field.Type = FIELD_FORM_TEXT
			case "textarea":
				field.Type = FIELD_FORM_TEXTAREA
			case "markdown":
				field.Type = FIELD_FORM_MARKDOWN
			case "editor":
				field.Type = FIELD_FORM_EDITOR
			case "mark":
				field.Type = FIELD_FORM_MARK
			case "path":
				field.Type = FIELD_FORM_PATH
			case "link":
				field.Type = FIELD_FORM_LINK
			case "email":
				field.Type = FIELD_FORM_EMAIL
			case "file":
				field.Type = FIELD_FORM_FILE
			case "time":
				field.Type = FIELD_FORM_TIME
			case "date":
				field.Type = FIELD_FORM_DATE
			case "int":
				field.Type = FIELD_FORM_INT
			case "float":
				field.Type = FIELD_FORM_FLOAT
			case "money":
				field.Type = FIELD_FORM_MONEY
			case "password":
				field.Type = FIELD_FORM_PASSWORD
			case "passwordt":
				field.Type = FIELD_FORM_PASSWORD_TWO
			case "enums":
				field.Type = FIELD_FORM_ENUM_STRING
			case "enumi":
				field.Type = FIELD_FORM_ENUM_INT
			case "enumf":
				field.Type = FIELD_FORM_ENUM_FLOAT
			default:
				field.Type = FIELD_FORM_OTHER
			}
		case 3:
			//是否可以为空
			if ocf == "true" {
				field.CanNull = true
			} else {
				field.CanNull = false
				continue
			}
		case 4:
			field.Info = ocf
		case 5:
			//最小值
			field.Min, _ = strconv.ParseInt(ocf, 10, 64)
		case 6:
			//最大值
			field.Max, _ = strconv.ParseInt(ocf, 10, 64)
		}
	}
}

type FormData struct {
	FieldsConfig AllFieldsConfig
	R            *http.Request
	Ip           *pubfunc.InputProcessor
}
type OneFormDataReturn struct {
	Int    int64
	Float  float64
	String string
	Bool   bool
}

func NewFormData(config *cpool.Section, r *http.Request) (fd *FormData) {
	afc := GetFieldConfig(config)
	fd = &FormData{FieldsConfig: afc, R: r, Ip: pubfunc.NewInputProcessor()}
	if fd.R.PostForm == nil {
		fd.R.ParseMultipartForm(defaultMaxMemory)
	}
	return
}

func (fd *FormData) GetAll() (all map[string]OneFormDataReturn, check bool, checks string) {
	all = make(map[string]OneFormDataReturn)
	for key, field := range fd.FieldsConfig {
		one := OneFormDataReturn{Int: 0, Float: 0, String: "", Bool: false}
		if field.UseIt == false {
			all[key] = one
			continue
		} else {
			check, checks = fd.checkOne(field, &one, fd.R)
			if check == false {
				return
			} else {
				all[key] = one
			}
		}
	}
	return
}

func (fd *FormData) checkOne(fc *FieldConfig, now *OneFormDataReturn, r *http.Request) (check bool, checks string) {
	check = false
	var ec int
	formd := r.PostForm[fc.DatabaseField]
	if len(formd) > 0 {
		switch fc.Type {
		case FIELD_FORM_TEXT:
			now.String, ec = fd.Ip.Text(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_TEXTAREA:
			now.String, ec = fd.Ip.Text(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_MARKDOWN:
			now.String, ec = fd.Ip.EditorIn(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_EDITOR:
			now.String, ec = fd.Ip.EditorIn(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_MARK:
			now.String, ec = fd.Ip.Mark(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_PATH:
			now.String, ec = fd.Ip.Text(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_LINK:
			now.String, ec = fd.Ip.Url(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_EMAIL:
			now.String, ec = fd.Ip.Email(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_FILE:
			now.String, ec = fd.Ip.Text(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_TIME:
			now.Int, ec = fd.Ip.Int(formd[0], fc.CanNull, 0, 9999999999)
		case FIELD_FORM_DATE:
			now.Int, ec = fd.Ip.Int(formd[0], fc.CanNull, 0, 9999999999)
		case FIELD_FORM_INT:
			now.Int, ec = fd.Ip.Int(formd[0], fc.CanNull, fc.Min, fc.Max)
		case FIELD_FORM_FLOAT:
			now.Float, ec = fd.Ip.Float(formd[0], fc.CanNull, float64(fc.Min), float64(fc.Max))
		case FIELD_FORM_MONEY:
			now.Float, ec = fd.Ip.Float(formd[0], fc.CanNull, float64(fc.Min), float64(fc.Max))
		case FIELD_FORM_PASSWORD:
			now.String, ec = fd.Ip.Password(formd[0], fc.CanNull)
		case FIELD_FORM_PASSWORD_TWO:
			now.String, ec = fd.Ip.PasswordTwo(formd[0], formd[1], fc.CanNull)
		}
		if ec != 0 {
			checks = fc.DatabaseField + " fields data not ok"
			return
		} else {
			check = true
			return
		}
	} else {
		checks = "fields not from formdata: " + fc.DatabaseField
		return
	}
}
