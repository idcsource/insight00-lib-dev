// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package pubfunc

import (
	"strings"

	"github.com/idcsource/Insight-0-0-lib/random"
)

// 将提供的字符串进行拆分词语处理
func SplitWords(str string) (normal [][]string) {
	strn := []rune(str)
	var tmpstring string
	tempslice := make([]string, 0)
	for _, one := range strn {
		// 碰到段落就新建一个切片
		if subsection(one) == true {
			if len(tmpstring) > 0 {
				tmpstring = optDot(tmpstring)
				tempslice = append(tempslice, tmpstring)
				tmpstring = ""
			}
			if len(tempslice) > 0 {
				normal = append(normal, tempslice)
				tempslice = make([]string, 0)
			}
			continue
		}
		// 如果碰到的是单字节的字，并且不是空格，就加入临时字符串
		if len([]byte(string(one))) == 1 && string(one) != " " {
			tmpstring += string(one)
			continue
		}
		// 如果碰到空格，如果临时字符串里有东西，则加入临时切片
		if string(one) == " " || string(one) == "\n" || string(one) == "\r" || string(one) == "\t" {
			if len(tmpstring) > 0 {
				tmpstring = optDot(tmpstring)
				tempslice = append(tempslice, tmpstring)
				tmpstring = ""
			}
			continue
		}
		// 普通的字符，就直接加入临时切片
		tempslice = append(tempslice, string(one))
	}
	if len(tmpstring) > 0 {
		tempslice = append(tempslice, tmpstring)
	}
	if len(tempslice) > 0 {
		normal = append(normal, tempslice)
	}
	return
}

// 处理最后的标点
func optDot(str string) string {
	str = strings.Trim(str, ".")
	str = strings.Trim(str, ",")
	str = strings.Trim(str, "!")
	return str
}

// 划分段落
func subsection(str rune) bool {
	pun := []string{
		"。",
		"　",
		"·",
		"，",
		"！",
		"；",
		";",
		"？",
		"：",
		"、",
		"“",
		"”",
		"\"",
		"'",
		"<",
		">",
		"《",
		"》",
		"(",
		")",
		"（",
		"）",
		"…",
		"}",
		"{",
		"\n",
		"\r",
		"\t",
	}
	strs := string(str)
	for _, one := range pun {
		if one == strs {
			return true
		}
	}
	return false
}

func strpos(str, substr string) int {
	// 子串在字符串的字节位置
	result := strings.Index(str, substr)
	if result >= 0 {
		// 获得子串之前的字符串并转换成[]byte
		prefix := []byte(str)[0:result]
		// 将子串之前的字符串转换成[]rune
		rs := []rune(string(prefix))
		// 获得子串之前的字符串的长度，便是子串在字符串的字符位置
		result = len(rs)
	}

	return result
}

// EncodeSalt 加盐
func EncodeSalt(str, stream, salt string) string {
	salt = random.GetSha1Sum(salt)
	tmpStream := ""
	lockLen := len(stream)
	j := 0
	k := 0
	streamb := []byte(stream)
	for i := 0; i < len(str); i++ {
		if k == len(salt) {
			k = 0
		}
		strb := []byte(str)
		stri := strb[i]
		saltb := []byte(salt)
		saltk := saltb[k]
		j = (strpos(stream, string(stri)) + int(saltk)) % (lockLen)

		streamj := streamb[j]
		tmpStream += string(streamj)
		k++
	}
	return tmpStream
}

// DecodeSalt 解盐
func DecodeSalt(str, stream, salt string) string {
	salt = random.GetSha1Sum(salt)
	tmpStream := ""
	lockLen := len(stream)
	j := 0
	k := 0
	streamb := []byte(stream)
	for i := 0; i < len(str); i++ {
		if k == len(salt) {
			k = 0
		}
		strb := []byte(str)
		stri := strb[i]
		saltb := []byte(salt)
		saltk := saltb[k]
		j = strpos(stream, string(stri)) - int(saltk)
		for j < 0 {
			j = j + lockLen
		}

		streamj := streamb[j]
		tmpStream += string(streamj)
		k++
	}
	return tmpStream
}
