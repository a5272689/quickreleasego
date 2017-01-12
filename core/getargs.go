package core

import (
	"github.com/docopt/docopt-go"
	"strconv"
	"github.com/Unknwon/goconfig"
)

func Getargs() (map[string]interface{}) {
	a:=`usage: quickreleasego [-h] -H HOSTNAME -j JOB -v VERSION [-b MONITBIN]
                       [-d APPDIR] [-k PACKAGESDIR] [-u USER] [-g GREP]
                       [-p PORT] [-t TIMEOUT] [-T TYPE] [-U UDP] [-c CONF]

Options:
  -h, --help            查看帮助
  -H HOSTNAME, --hostname=HOSTNAME 服务器名，多个用‘,’号隔开
  -j JOB, --job=JOB     JOB_NAME变量,必须字段
  -v VERSION, --version=VERSION    变量,由平台传入,必须字段
  -b MONITBIN, --monitbin=MONITBIN      monit程序的bin文件，可以为空[default: /application/monit_py/bin/monit.py]
  -d APPDIR, --appdir=APPDIR     程序所在根目录，可以为空[default: /home/appdeploy/deploy/apps]
  -k PACKAGESDIR, --packagesdir=PACKAGESDIR     包所在根目录，可以为空[default: /home/appdeploy/deploy/packages]
  -u USER, --user=USER  app目录的所属用户，可以为空[default: appdeploy]
  -g GREP, --grep=GREP  程序关键词，可以为空
  -p PORT, --port=PORT  程序端口，可以为空
  -t TIMEOUT, --timeout=TIMEOUT  程序端口，可以为空[default: 30]
  -T TYPE, --type=TYPE  程序发布方式，可以为空[default: common],他可以是reload
  -U UDP, --udp=UDP     程序端口，可以为空
  -c CONF, --conf=conf     配置文件[default: /etc/quickreleasego.ini]
	`
	arguments, _ := docopt.Parse(a, nil, true, "", false)
	newargs:=make(map[string] interface{})
	for key:=range arguments{
		newkey:=key[2:]
		newargs[newkey]=arguments[key]
	}
	if newargs["timeout"]!=nil{
		newtimeout,_:=strconv.Atoi(newargs["timeout"].(string))
		newargs["timeout"]=newtimeout
	}else {
		newargs["timeout"]=0
	}
	if newargs["port"]!=nil{
		port,_:=strconv.Atoi(newargs["port"].(string))
		newargs["port"]=port
	}else {
		newargs["port"]=0
	}
	if newargs["udp"]=="true"{
		newargs["udp"]=true
	}else {
		newargs["udp"]=false
	}
	conf,confEer:=goconfig.LoadConfigFile(newargs["conf"].(string))
	if confEer==nil{
		fileserver_url,_:=conf.GetValue("DEFAULT","fileserver_url")
		saltapi_url,_:=conf.GetValue("DEFAULT","saltapi_url")
		newargs["fileserver_url"]=fileserver_url
		newargs["saltapi_url"]=saltapi_url
	}else {
		panic(confEer)
	}
	return newargs
}
