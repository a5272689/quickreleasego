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
}

func (self Release) Commonrelease(hostname string,c chan map[string]interface{}) {
	jid_jsmap:=make(map[string]interface{})
	defer func() {
		funcErr:=recover()
		if funcErr!=nil{
			c <- jid_jsmap
		}
		close(c)
	}()
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
		jid_jsmap=self.usereleaseAPI(hostname,urlpath,v)
	}
	c <- jid_jsmap
}

func (self Release) usereleaseAPI(hostname,urlpath string,v url.Values) (map[string]interface{}) {
	jid_jsmap:=make(map[string]interface{})
	resp, err := http.PostForm(urlpath,v)
	if err != nil {
		jid_jsmap[hostname]=fmt.Sprintf("调用接口失败！！,错误信息：（%v）",err)
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
	return jid_jsmap
}

func (self Release) bodyjsontomap(body,hostname,urlpath string) (map[string]interface{}) {
	js,_:=simplejson.NewFromReader(strings.NewReader(body))
	jsmap,_:=js.Map()
	if len(jsmap)==0{
		tmp_jsmap:=make(map[string]interface{})
		tmp_jsmap[hostname]=fmt.Sprintf("调用接口（%v）失败！！,错误信息：（%v）",urlpath,body)
		jsmap=tmp_jsmap
	}
	return jsmap
}


func (self Release) Reloadrelease(hostname string,c chan map[string]interface{}){
	jid_jsmap:=make(map[string]interface{})
	defer func() {
		funcErr:=recover()
		if funcErr!=nil{
			c <- jid_jsmap
		}
		close(c)
	}()
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
		jid_jsmap=self.usereleaseAPI(hostname,urlpath,v)
	}
	c <- jid_jsmap
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
	var chan_getpackages []chan map[string]interface{}
	var chan_jids []chan map[string]interface{}
	var chan_jid_results []chan map[string]interface{}
	for ind:=range self.Args["hostname"].([]string){
		getpackage_chan:=make(chan map[string]interface{})
		go self.getpackage(self.Args["hostname"].([]string)[ind],getpackage_chan)
		chan_getpackages=append(chan_getpackages,getpackage_chan)
	}
	for _,tmp_chan:=range chan_getpackages{
		for hostname,result:=range <- tmp_chan{
			if result.(string)!=self.Args["packagespath"]{
				fmt.Println("主机： ",hostname," 获取包失败：",result)
			}else {
				fmt.Println("主机： ",hostname," 获取包成功，包存放路径：",result)
				release_jid_chan:=make(chan map[string]interface{})
				chan_jids=append(chan_jids,release_jid_chan)
				if self.Args["type"]=="common"{
					go self.Commonrelease(hostname,release_jid_chan)
				}else if self.Args["type"]=="reload"{
					go self.Reloadrelease(hostname,release_jid_chan)
				}else {
					fmt.Println(fmt.Sprintf("主机(%v)发布失败！！，发布方式(%v)不存在！！",hostname,self.Args["type"]))
					close(release_jid_chan)
				}
			}
		}
	}
	for _,tmp_chan:=range chan_jids{
		for hostname,jid:=range <- tmp_chan{
				if jid.(map[string]interface {})["jid"]==nil{
					fmt.Println(fmt.Sprintf("主机(%v)发布失败！！，发布方式(%v)的接口调用失败！！,调用结果：%v",hostname,self.Args["type"],jid))
				}else {
					release_result_chan:=make(chan map[string]interface{})
					go self.getjidresult(hostname,jid.(map[string]interface {})["jid"].(string),release_result_chan)
					chan_jid_results=append(chan_jid_results,release_result_chan)
				}
		}
	}
	for _,tmp_chan:=range chan_jid_results{
		for hostname,result:=range <-tmp_chan{
			if fmt.Sprint(reflect.TypeOf(result.(map[string]interface {})["ret"]))=="string"{
				fmt.Println("主机：",hostname," 发布失败！！错误信息：",result.(map[string]interface {})["ret"])
			}else {
				ret:=result.(map[string]interface {})["ret"].(map[string]interface {})
				if ret["result"].(bool){
					fmt.Println("主机：",hostname,"发布成功！！ 发布信息：",ret["info"].(string))
				}else {
					fmt.Println("主机：",hostname,"发布失败！！ 发布信息：",ret["info"].(string))
				}
			}
		}
	}

}

func (self Release) getjidresult(hostname,jid string,c chan map[string]interface{})  {
	result:=make(map[string]interface{})
	defer func() {
		funcErr:=recover()
		if funcErr!=nil{
			c <- result
		}
		close(c)
	}()
	index:=0
	for index==0{
		urlpath := fmt.Sprintf("%v/saltAPI/async", self.Args["saltapi_url"])
		v := url.Values{}
		v.Set("jid", jid)
		resp, err := http.PostForm(urlpath,v)
		if err != nil {
			result[hostname]=fmt.Sprintf("调用接口失败！！,错误信息：（%v）",err)
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
			result=jsmap
			c <- result
			index=1
		}else {
			time.Sleep(time.Second*1)

		}
	}
}

func (self Release) getpackage(hostname string,c chan map[string]interface{})  {
	result:=make(map[string]interface{})
	defer func() {
		funcErr:=recover()
		if funcErr!=nil{
			c <- result
		}
		close(c)
	}()
	urlpath := fmt.Sprintf("%v/saltAPI", self.Args["saltapi_url"])
	v := url.Values{}
	v.Set("tgt", hostname)
	v.Set("fun", "releasefj.get_package_url")
	v.Set("arg", self.Args["packageurl"].(string))
	v.Add("arg", self.Args["packagespath"].(string))
	resp, err := http.PostForm(urlpath, v)

	if err == nil {
		body,_ := ioutil.ReadAll(resp.Body)
		result=self.bodyjsontomap(string(body),hostname,urlpath)
	}else {
		result[hostname]=fmt.Sprintf("调用接口失败！！,错误信息：（%v）",err)
	}
	defer resp.Body.Close()
	c<- result
}
