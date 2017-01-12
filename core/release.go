package core

import (
	"fmt"
	"strings"
	"path"
	"net/http"
	"net/url"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
)

type Release struct {Args map[string] interface{}}

func (self Release) Commonrelease() (result bool,info string) {
	fmt.Println(self.Args)

	return true,"common"
}

func (self Release) Reloadrelease() (result bool,info string) {
	fmt.Println(self.Args)

	return true,"reload"
}

func (self Release) Call() (result bool,info string)  {
	self.Args["hostname"]=strings.Split(self.Args["hostname"].(string),",")
	self.Args["apppath"]=path.Join(self.Args["appdir"].(string),self.Args["job"].(string))
	tmppackagepath:=path.Join(self.Args["job"].(string),fmt.Sprintf("%v_%v.tar.gz",self.Args["job"],self.Args["version"]))
	self.Args["packageurl"]=self.Args["fileserver_url"].(string)+path.Join("/packages",tmppackagepath)
	self.Args["packagespath"]=path.Join(self.Args["packagesdir"].(string),tmppackagepath)
	fmt.Println(self.Args["packageurl"])
	self.GetPackage()
	if self.Args["type"]=="common"{
		return self.Commonrelease()
	}else if self.Args["type"]=="reload"{
		return self.Reloadrelease()
	}else {
		return false,"发布方式不存在！！"
	}

}

func (self Release) GetPackage() {
	urlpath := fmt.Sprintf("%v/saltAPI", self.Args["saltapi_url"])
	fmt.Println(urlpath)
	cs := make(chan map[string]interface{})
	for ind:=range self.Args["hostname"].([]string){
		go getpackage(urlpath,self.Args["packageurl"].(string),self.Args["packagespath"].(string),self.Args["hostname"].([]string)[ind],cs)
	}
	num:=len(self.Args["hostname"].([]string))
	init_num:=0
	for i := range cs {
		init_num+=1
        	fmt.Println(i)
		if init_num==num{
			close(cs)
		}
    	}
}

func getpackage(urlpath,packageurl,packagespath,hostname string,cs chan map[string]interface{})  {
	v := url.Values{}
	v.Set("tgt", hostname)
	v.Set("fun", "releasefj.get_package_url")
	v.Set("arg", packageurl)
	v.Add("arg", packagespath)
	resp, err := http.PostForm(urlpath, v)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tmpbody:=[]byte("{}")
		body=tmpbody
	}
	js,_:=simplejson.NewFromReader(strings.NewReader(string(body)))
	jsmap,_:=js.Map()
	if jsmap[hostname]==nil{
		jsmap[hostname]=fmt.Sprintf("主机：%v 不存在或已经下线！！",hostname)
	}
	cs <- jsmap
}

