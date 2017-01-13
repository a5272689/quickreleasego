package core

import (
	"fmt"
	"strings"
	"path"
	"net/http"
	"net/url"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"time"
	"reflect"
)

type Release struct {
	Args map[string] interface{}
	chan_getpackage_result chan map[string]interface{}
	chan_release_result chan map[string]interface{}
	chan_jid_result chan map[string]interface{}
}

func (self Release) Commonrelease(hostname string) {
	jid_jsmap:=make(map[string]interface{})
	if self.Args["grep"]==nil{
		jid_jsmap[hostname]=make(map[string]interface{})
	}else {
		urlpath := fmt.Sprintf("%v/saltAPI/async", self.Args["saltapi_url"])
		js:=simplejson.New()
		js.Set("startuser",self.Args["user"])
		js.Set("app_path",self.Args["apppath"])
		js.Set("pgrep",self.Args["grep"])
		js.Set("version",self.Args["version"])
		js.Set("packages_path",self.Args["packagesdir"])
		js.Set("timeout",self.Args["timeout"])
		js.Set("udp",self.Args["udp"])
		js.Set("monit_bin",self.Args["monitbin"])
		js_byte,_:=js.MarshalJSON()
		js_str:=string(js_byte)
		v := url.Values{}
		v.Set("tgt", hostname)
		v.Set("fun", "releasefj.release_package")
		v.Set("arg", js_str)
		resp, err := http.PostForm(urlpath,v)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tmpbody:=[]byte("{}")
			body=tmpbody
		}
		js_jid,_:=simplejson.NewFromReader(strings.NewReader(string(body)))
		jsmap,_:=js_jid.Map()
		jid_jsmap[hostname]=jsmap
	}

	self.chan_release_result <- jid_jsmap
}

func (self Release) bodyjsontomap(body,hostname string) (map[string]interface{}) {
	js,_:=simplejson.NewFromReader(strings.NewReader(body))
	jsmap,_:=js.Map()
	if jsmap[hostname]==nil{
		if len(jsmap)>=1{
			tmp_jsmap:=make(map[string]interface{})
			tmp_jsmap[hostname]=fmt.Sprintln(jsmap)
			jsmap=tmp_jsmap
		}else {
			jsmap[hostname]=fmt.Sprintf("主机：%v 不存在或已经下线！！",hostname)
		}
	}
	return jsmap
}


func (self Release) Reloadrelease(hostname string){
	jid_jsmap:=make(map[string]interface{})
	if self.Args["port"]==nil||self.Args["port"]==0{
		jid_jsmap[hostname]=make(map[string]interface{})
	}else {
		urlpath := fmt.Sprintf("%v/releasePR", self.Args["saltapi_url"])
		js:=simplejson.New()
		js.Set("startuser",self.Args["user"])
		js.Set("app_path",self.Args["apppath"])
		js.Set("port",self.Args["port"])
		js.Set("version",self.Args["version"])
		js.Set("packages_path",self.Args["packagesdir"])
		js.Set("timeout",self.Args["timeout"])
		js.Set("udp",self.Args["udp"])
		js.Set("monit_bin",self.Args["monitbin"])
		js2:=simplejson.New()
		js2.Set("kwargs",js)
		js2.Set("callback_url","http://127.0.0.1")
		js2.Set("tgt",hostname)
		js_byte,_:=js2.MarshalJSON()
		js_str:=string(js_byte)
		v := url.Values{}
		v.Set("data", js_str)
		resp, err := http.PostForm(urlpath,v)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tmpbody:=[]byte("{}")
			body=tmpbody
		}
		js_jid,_:=simplejson.NewFromReader(strings.NewReader(string(body)))
		jsmap,_:=js_jid.Map()
		jid_jsmap[hostname]=jsmap
	}

	self.chan_release_result <- jid_jsmap
}

func (self Release) Call() {
	self.Args["hostname"]=strings.Split(self.Args["hostname"].(string),",")
	self.Args["apppath"]=path.Join(self.Args["appdir"].(string),self.Args["job"].(string))
	tmppackagepath:=path.Join(self.Args["job"].(string),fmt.Sprintf("%v_%v.tar.gz",self.Args["job"],self.Args["version"]))
	self.Args["packageurl"]=self.Args["fileserver_url"].(string)+path.Join("/packages",tmppackagepath)
	self.Args["packagespath"]=path.Join(self.Args["packagesdir"].(string),"packages",tmppackagepath)
	self.GetPackageRelease()
}

func (self Release) GetPackageRelease() {
	self.chan_getpackage_result = make(chan map[string]interface{})
	self.chan_release_result = make(chan map[string]interface{})
	self.chan_jid_result = make(chan map[string]interface{})
	for ind:=range self.Args["hostname"].([]string){
		go self.getpackage(self.Args["hostname"].([]string)[ind])
	}
	num:=len(self.Args["hostname"].([]string))
	init_num:=0
	releasecount:=0
	for getpackage_result := range self.chan_getpackage_result {
		init_num+=1
        	for hostname:=range getpackage_result{
			if getpackage_result[hostname]==self.Args["packagespath"]{
				fmt.Println(fmt.Sprintf("主机(%v)分发包成功！！，分发包存储路径：%v",hostname,getpackage_result[hostname]))
				releasecount+=1
				if self.Args["type"]=="common"{
					go self.Commonrelease(hostname)
				}else if self.Args["type"]=="reload"{
					go self.Reloadrelease(hostname)
				}else {
					releasecount-=1
					fmt.Println(fmt.Sprintf("主机(%v)发布失败！！，发布方式(%v)不存在！！",hostname,self.Args["type"]))
				}
			}else {
				fmt.Println(fmt.Sprintf("主机(%v)分发包失败！！，saltAPI返回信息：%v",hostname,getpackage_result[hostname]))
			}
		}
		if init_num==num{
			close(self.chan_getpackage_result)
		}
    	}
	init_num=0
	jidcount:=0
	if releasecount>0{
		for releaserelust:=range self.chan_release_result{
			init_num+=1

			for hostname:=range releaserelust{
				if releaserelust[hostname].(map[string]interface {})["jid"]==nil{
					fmt.Println(fmt.Sprintf("主机(%v)发布失败！！，发布方式(%v)的接口调用失败！！,调用信息：%v",hostname,self.Args["type"],releaserelust))
				}else {
					jidcount+=1
					go self.getjidresult(hostname,releaserelust[hostname].(map[string]interface {})["jid"].(string))
				}
			}
			if init_num==releasecount{
				close(self.chan_release_result)
			}
		}
	}

	init_num=0
	if jidcount>0{
		for jidrelust:=range self.chan_jid_result{
			init_num+=1
			for hostname:=range jidrelust{
				if fmt.Sprintln(reflect.TypeOf(jidrelust[hostname].(map[string]interface {})["ret"]))=="string\n"{
					fmt.Println(jidrelust[hostname].(map[string]interface {})["ret"])
				}else {
					ret:=jidrelust[hostname].(map[string]interface {})["ret"].(map[string]interface {})
					if ret["result"].(bool){
						fmt.Println("发布成功！！ 发布信息：",ret["info"].(string))
					}else {
						fmt.Println("发布失败！！ 发布信息：",ret["info"].(string))
					}
				}
			}
			if init_num==jidcount{
				close(self.chan_jid_result)
			}

		}
	}

}

func (self Release) getjidresult(hostname,jid string)  {
	index:=0
	for index==0{
		urlpath := fmt.Sprintf("%v/saltAPI/async", self.Args["saltapi_url"])
		v := url.Values{}
		v.Set("jid", jid)
		resp, err := http.PostForm(urlpath,v)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tmpbody:=[]byte("{}")
			body=tmpbody
		}
		js_jid,_:=simplejson.NewFromReader(strings.NewReader(string(body)))
		jsmap,_:=js_jid.Map()
		if jsmap["jid"]==nil{
			if jsmap[hostname]==nil{
				jsmap[hostname]=map[string]interface {} {"ret":"jid查询接口调用失败!!"}
			}else if fmt.Sprintln(reflect.TypeOf(jsmap[hostname]))!="map[string]interface {}\n"{
				jsmap[hostname]=map[string]interface {} {"ret":jsmap[hostname].(string)}
			}
			self.chan_jid_result <- jsmap
			index=1
		}else {
			time.Sleep(5)

		}
	}
}

func (self Release) getpackage(hostname string)  {
	urlpath := fmt.Sprintf("%v/saltAPI", self.Args["saltapi_url"])
	v := url.Values{}
	v.Set("tgt", hostname)
	v.Set("fun", "releasefj.get_package_url")
	v.Set("arg", self.Args["packageurl"].(string))
	v.Add("arg", self.Args["packagespath"].(string))
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
	jsmap:=self.bodyjsontomap(string(body),hostname)
	self.chan_getpackage_result <- jsmap
}
