// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"github.com/saintfish/chardet"

	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

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

func charEncodeToUtf8(body []byte) (string, error) {
	cd := chardet.NewHtmlDetector()
	ccode, err := cd.DetectBest(body)
	if err != nil {
		return "", err
	}
	var bodys string
	if ccode.Charset != "UTF-8" {
		thecode := ccode.Charset
		if strings.Contains(thecode, "GB-") {
			thecode = "GBK"
		}
		var err2 error
		bodys, err2 = iconv.ConvertString(string(body), thecode, "utf-8")
		if err2 != nil {
			return "", err2
		}
	} else {
		bodys = string(body)
	}
	return bodys, nil
}

func trimHtml(html string) string {
	html = strings.TrimSpace(html)
	html = strings.Replace(html, "\r", "\n", -1)

	b0, _ := regexp.Compile(`\&[^ ]{1,6};`)
	html = b0.ReplaceAllString(html, " ")

	b1, _ := regexp.Compile(`(?isU)<HEAD>(.*)</head>`)
	html = b1.ReplaceAllString(html, "\n")

	b4, _ := regexp.Compile(`(?isU)<script(.*)</script>`)
	html = b4.ReplaceAllString(html, "\n")
	b6, _ := regexp.Compile(`(?isU)<style(.*)</style>`)
	html = b6.ReplaceAllString(html, "\n")
	//b2, _ := regexp.Compile("<(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b2, _ := regexp.Compile("<([^>]+)>")
	//b3, _ := regexp.Compile("</(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b3, _ := regexp.Compile("</([^>]+)>")
	html = b2.ReplaceAllString(html, "\n")
	html = b3.ReplaceAllString(html, "\n")

	//html = strings.Replace(html,"\t"," ",-1);
	//html = strings.Replace(html,"\n"," ",-1);
	//html = strings.Replace(html, "-", " ", -1)
	//html = strings.Replace(html, "/", " ", -1)
	//html = strings.Replace(html, "\\", " ", -1)
	//html = strings.Replace(html, ":", " ", -1)

	b5, _ := regexp.Compile("([ ]{2,})")
	html = b5.ReplaceAllString(html, " ")
	b7, _ := regexp.Compile(`([\t]{2,})`)
	html = b7.ReplaceAllString(html, "\n")
	b8, _ := regexp.Compile(`([　]{2,})`)
	html = b8.ReplaceAllString(html, " ")
	b9, _ := regexp.Compile(`([ \n]{2,})`)
	html = b9.ReplaceAllString(html, "\n")

	b5_1, _ := regexp.Compile(`([\n]{2,})`)
	html = b5_1.ReplaceAllString(html, "\n")
	html = strings.ToLower(html)
	return html
}

func getAllUrl(htmlbody string, fatherurl string) (urls []UrlBasic, err error) {
	fatherpre, _ := url.Parse(fatherurl)
	bodysreader := strings.NewReader(htmlbody)
	document, err := goquery.NewDocumentFromReader(bodysreader)
	if err != nil {
		fmt.Println(err)
		return
	}
	urls = make([]UrlBasic, 0)
	document.Find("a").Each(func(i int, se *goquery.Selection) {
		link, exist := se.Attr("href")
		if exist == false {
			return
		}
		linka, err := url.Parse(link)
		if err != nil {
			return
		}
		linkComplete(linka, fatherpre)
		url := UrlBasic{
			Url:    linka.String(),
			Text:   se.Text(),
			Domain: linka.Hostname(),
		}
		urls = append(urls, url)
	})
	return
}

func getTitle(html string) string {
	bodysreader := strings.NewReader(html)
	jquery, _ := goquery.NewDocumentFromReader(bodysreader)
	title := jquery.Find("title").Text()
	return title
}

func getKeyword(html string) []string {
	bodysreader := strings.NewReader(html)
	jquery, _ := goquery.NewDocumentFromReader(bodysreader)
	keyword, _ := jquery.Find("meta[name=keywords]").Attr("content")
	keywords := strings.Split(keyword, ",")
	if len(keywords) == 1 {
		keywords = strings.Split(keyword, ";")
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, "，")
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, " ")
	}
	return keywords
}

func linkComplete(linka, father *url.URL) {
	if linka.Host == "" || linka.Scheme == "" {
		linka.Host = father.Host
		linka.Scheme = father.Scheme
		abpath, _ := regexp.MatchString("^/", linka.Path)
		if abpath == false {
			linka.Path = pubfunc.DirMustEnd(father.Path) + linka.Path
		}
	}
	return
}
