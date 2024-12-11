package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed editormd/*
var editormdFS embed.FS

var _hugo string
var _port string
var _home string
var _post string
var _passwd string

func init() {
	flag.StringVar(&_passwd, "passwd", "", "接口加密密码")
	flag.StringVar(&_port, "port", "58880", "监听端口")
	flag.StringVar(&_hugo, "hugo", "./hugo.exe", "hugo可执行程序的路径")
	flag.StringVar(&_home, "home", "./blog", "创建的Blog根目录")
	flag.StringVar(&_post, "posts", "content/posts", "创建的Blog文章目录")
	flag.Parse()

	_hugo, _ = filepath.Abs(_hugo)
	_home, _ = filepath.Abs(_home)

	_, err := os.Stat(_home)
	if os.IsNotExist(err) {
		os.MkdirAll(_home, os.ModePerm)
	}

	_, err = os.Stat(_home + "/" + _post)
	if os.IsNotExist(err) {
		os.MkdirAll(_home+"/"+_post, os.ModePerm)
	}
}

type HttpBack struct {
	Code int         `json:code`
	Info interface{} `json:info`
}

func check(passwd string) bool {
	// 未设置密码则不进行校验
	if _passwd == "" {
		return true
	}

	pw1 := base64.StdEncoding.EncodeToString([]byte(passwd))
	pw2 := base64.StdEncoding.EncodeToString([]byte(_passwd))

	return pw1 == pw2
}

func IsCheck(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack
	var pass string = ""

	_, check1 := r.Form["passwd"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info = (_passwd != "")
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 创建新文章
func MdNew(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack
	var name string = ""
	var pass string = ""

	_, check1 := r.Form["passwd"]
	_, check2 := r.Form["name"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if check2 {
		name = r.Form["name"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	tmpFilePath := fmt.Sprintf("%s/%s/%s", _home, _post, name)
	_, err := os.Stat(tmpFilePath)
	if os.IsExist(err) || err == nil {
		result.Code = 10000
		result.Info = "posts existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	// 创建默认的文章
	cmd := exec.Command(_hugo, "new", fmt.Sprintf("%s/%s.md", _post, name))
	cmd.Dir = _home
	err = cmd.Run()
	if err != nil {
		result.Code = 10001
		result.Info = fmt.Sprintf("failed with %s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	// 修改文章目录结构
	// 创建目录 xxx/[文章名]
	err = os.MkdirAll(tmpFilePath, os.ModePerm)
	if err != nil {
		result.Code = 10002
		result.Info = fmt.Sprintf("failed with %s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}
	// 创建插图目录
	err = os.MkdirAll(fmt.Sprintf("%s/%s", tmpFilePath, "img"), os.ModePerm)
	if err != nil {
		result.Code = 10003
		result.Info = fmt.Sprintf("failed with %s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}
	// 将xxx/[文章名].md 移动并改名 xxx/[文章名]/index.md
	err = os.Rename(tmpFilePath+".md", tmpFilePath+"/index.md")
	if err != nil {
		result.Code = 10004
		result.Info = fmt.Sprintf("failed with %s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info, _ = os.ReadFile(tmpFilePath + "/index.md")
	// fmt.Printf("%s", result.Info)
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 保存文章内容
func MdSave(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack

	var name string = ""
	var pass string = ""
	var content string = ""

	_, check1 := r.Form["passwd"]
	_, check2 := r.Form["name"]
	_, check3 := r.Form["content"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if check2 {
		name = r.Form["name"][0]
	}

	if check3 {
		content = r.Form["content"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	tmpFilePath := fmt.Sprintf("%s/%s/%s", _home, _post, name)
	_, err := os.Stat(tmpFilePath)
	if os.IsNotExist(err) {
		result.Code = 20000
		result.Info = "posts not existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	var decodedBytes []byte
	decodedBytes, err = base64.StdEncoding.DecodeString(content)
	if err != nil {
		result.Code = 20001
		result.Info = fmt.Sprintf("base64 decode err:%s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	err = os.WriteFile(tmpFilePath+"/index.md", decodedBytes, os.ModePerm)
	if err != nil {
		result.Code = 20002
		result.Info = fmt.Sprintf("save file err:%s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info = content
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 获取文章列表
func MdList(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack

	// var pass string = ""

	// _, check1 := r.Form["passwd"]

	// if check1 {
	// 	pass = r.Form["passwd"][0]
	// }

	// if !check(pass) {
	// 	result.Code = 99999
	// 	result.Info = "wrong password."
	// 	bytes, _ := json.Marshal(result)
	// 	w.Write(bytes)
	// 	return
	// }

	root := fmt.Sprintf("%s/%s", _home, _post)

	filesDE, err := os.ReadDir(root)
	if err != nil {
		result.Code = 40000
		result.Info = "get posts list err:" + err.Error()
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	var files []string
	for _, file := range filesDE {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		files = append(files, file.Name())
	}

	if len(files) == 0 {
		result.Code = 40001
		result.Info = "posts list empty."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info = files
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 打开文章
func MdOpen(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack

	var name string = ""
	var pass string = ""

	_, check1 := r.Form["passwd"]
	_, check2 := r.Form["name"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if check2 {
		name = r.Form["name"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	tmpFilePath := fmt.Sprintf("%s/%s/%s", _home, _post, name)
	_, err := os.Stat(tmpFilePath)
	if os.IsNotExist(err) {
		result.Code = 40000
		result.Info = "posts not existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info, _ = os.ReadFile(tmpFilePath + "/index.md")
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 删除文章
func MdDel(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack

	var name string = ""
	var pass string = ""

	_, check1 := r.Form["passwd"]
	_, check2 := r.Form["name"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if check2 {
		name = r.Form["name"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	tmpFilePath := fmt.Sprintf("%s/%s/%s", _home, _post, name)
	_, err := os.Stat(tmpFilePath)
	if os.IsNotExist(err) {
		result.Code = 50000
		result.Info = "posts not existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	os.RemoveAll(tmpFilePath)
	os.RemoveAll(tmpFilePath + ".md")

	result.Code = 200
	result.Info = "posts has deleted."
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 生成博客资源
func Hugo(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "接口异常", 99999)
		}
	}()

	r.ParseForm()

	if r.Method == http.MethodOptions {
		// 1. [必须]接受指定域的请求，可以使用*不加以限制，但不安全
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// 2. [必须]设置服务器支持的所有跨域请求的方法
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		// 3. [可选]服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length,Token")
		// 4. [可选]设置XMLHttpRequest的响应对象能拿到的额外字段
		w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers,Token")
		// 5. [可选]是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		return
	}

	var result HttpBack

	var pass string = ""

	_, check1 := r.Form["passwd"]

	if check1 {
		pass = r.Form["passwd"][0]
	}

	if !check(pass) {
		result.Code = 99999
		result.Info = "wrong password."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	cmd := exec.Command(_hugo)
	cmd.Dir = _home
	err := cmd.Run()
	if err != nil {
		result.Code = 60000
		result.Info = fmt.Sprintf("failed with hugo:%s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info = "blog source gen success."
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

// 重定向主页
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/editormd/public", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/editor-helper", redirectHandler)
	http.HandleFunc("/new", MdNew)
	http.HandleFunc("/save", MdSave)
	http.HandleFunc("/list", MdList)
	http.HandleFunc("/open", MdOpen)
	http.HandleFunc("/del", MdDel)
	http.HandleFunc("/hugo", Hugo)
	http.HandleFunc("/ischeck", IsCheck)

	// 代理editor.md的静态资源
	http.Handle("/editormd/", http.FileServer(http.FS(editormdFS)))

	http.Handle("/blog/", http.StripPrefix("/blog/", http.FileServer(http.Dir(_home+"/public"))))

	fmt.Println("hugo管理助手已启动 端口:", _port)
	fmt.Printf("    主页地址:0.0.0.0:%s", _port)
	err := http.ListenAndServe(":"+_port, nil)
	if err != nil {
		fmt.Println("服务器开启错误: ", err)
	}
}
