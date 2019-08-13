package main

import (
	"fmt"
	"regexp"

	"github.com/astaxie/beego/httplib"
)

func main() {
	fmt.Println("hello world!")
	reqstr, _ := httplib.Get("http://sports.sina.com.cn/nba/").String()
	bytehtml := []byte(reqstr)
	// fmt.Println(req.String())
	//rule := `<a href="/celebrity/[0-9]+/" rel="v:directedBy">(.*)</a>`
	// 正则需要输出的字符串放在(.*)里，过滤/2019-08-12还需学习
	rule := `<a href="(.*)" target="_blank">`
	result := GetValue(rule, &bytehtml)
	fmt.Println(result)
}

func GetValue(rule string, sHtml *[]byte) string {
	reg := regexp.MustCompile(rule)
	result := reg.FindAllStringSubmatch(string(*sHtml), -1)
	if len(result) == 0 {
		fmt.Println("result is empty")
		return ""
	}
	if len(result[0]) == 0 {
		fmt.Println("result[0] is empty")
		return ""
	}
	fmt.Println(result)
	for i := 0; i < len(result); i++ {
		fmt.Println(result[i][1])
	}
	//fmt.Println(result[1][1])
	return result[0][1]
}
