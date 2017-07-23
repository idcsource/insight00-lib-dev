// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs

import (
	"strings"
	"regexp"
	
	"github.com/idcsource/insight00-lib/random"
)

func strpos(str,substr string) int {  
      // 子串在字符串的字节位置  
      result := strings.Index(str,substr)    
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
func EncodeSalt (str, stream, salt string) string {
	salt = random.GetSha1Sum(salt);
	tmpStream := "";
	lockLen := len(stream);
	j := 0;
	k := 0;
	streamb := []byte(stream);
	for i := 0; i < len(str); i++ {
		if k == len(salt) {
			k = 0;
		}
		strb := []byte(str);
		stri := strb[i];
		saltb := []byte(salt);
		saltk := saltb[k];
		j = (strpos(stream, string(stri)) + int(saltk))%(lockLen);
		
		streamj := streamb[j];
		tmpStream += string(streamj);
		k++;
	}
	return tmpStream;
}

// DecodeSalt 解盐
func DecodeSalt (str, stream, salt string) string {
	salt = random.GetSha1Sum(salt);
	tmpStream := "";
	lockLen := len(stream);
	j := 0;
	k := 0;
	streamb := []byte(stream);
	for i := 0; i < len(str); i++ {
		if k == len(salt) {
			k = 0;
		}
		strb := []byte(str);
		stri := strb[i];
		saltb := []byte(salt);
		saltk := saltb[k];
		j = strpos(stream, string(stri)) - int(saltk);
		for j < 0 {
			j = j + lockLen;
		}
		
		streamj := streamb[j];
		tmpStream += string(streamj);
		k++;
	}
	return tmpStream;
}



//将Url用斜线/拆分
// :id=dfad/:type=dafa/
func SplitUrl (url string) (urla []string, parameter map[string]string ) {
	parameter = make(map[string]string);
	urlRequest := strings.Split(url, "/");
	matchP, _ := regexp.Compile("^:([A-Za-z0-9_-]+)=(.*)");
	for _, v := range urlRequest{
		if ( len(v) != 0 ) {
			if matchP.MatchString(v) {
				pa := matchP.FindStringSubmatch(v);
				parameter[pa[1]] = pa[2];
			}else{
				urla = append(urla, v);
			}
		}
	}
	return;
}
