// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import "strings"

// let string to index char
func toSequence(str string) (normal map[uint64][]string) {
	var charcount uint64 = 0
	normal = make(map[uint64][]string)
	strn := []rune(str)
	var tmpstring string
	tempslice := make([]string, 0)
	for _, one := range strn {
		charcount++
		// if a new sentence, create a new slice
		if subSentences(one) == true {
			if len(tmpstring) > 0 {
				tmpstring = optDot(tmpstring)
				tempslice = append(tempslice, tmpstring)
				tmpstring = ""
			}
			if len(tempslice) > 0 {
				nowcount := charcount - 1 - uint64(len(tempslice))
				normal[nowcount] = tempslice
				tempslice = make([]string, 0)
			}
			continue
		}
		// if is a one byte char, and not a space, add it to temporary string
		if len([]byte(string(one))) == 1 && string(one) != " " {
			tmpstring += string(one)
			continue
		}
		// if is a space, and if the temporary string not null, add to temporary slice
		if string(one) == " " || string(one) == "\n" || string(one) == "\r" || string(one) == "\t" {
			if len(tmpstring) > 0 {
				tmpstring = optDot(tmpstring)
				tempslice = append(tempslice, tmpstring)
				tmpstring = ""
			}
			continue
		}
		// normal char, add to temporary slice
		tempslice = append(tempslice, string(one))
	}
	if len(tmpstring) > 0 {
		tempslice = append(tempslice, tmpstring)
	}
	if len(tempslice) > 0 {
		nowcount := charcount - 1 - uint64(len(tempslice))
		normal[nowcount] = tempslice
	}
	return
}

func optDot(str string) string {
	str = strings.Trim(str, ".")
	str = strings.Trim(str, ",")
	str = strings.Trim(str, "!")
	return str
}

func subSentences(str rune) bool {
	if str <= 31 {
		return true
	}
	if str >= 33 && str <= 47 {
		return true
	}
	if str >= 58 && str <= 64 {
		return true
	}
	if str >= 91 && str <= 96 {
		return true
	}
	if str >= 123 && str <= 127 {
		return true
	}
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
		"|",
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
