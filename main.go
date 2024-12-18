package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

//go:embed editormd/*
var editormdFS embed.FS

var _hugo string
var _ip string
var _port string
var _home string
var _post string
var _passwd string

func init() {
	flag.StringVar(&_passwd, "passwd", "", "接口加密密码")
	flag.StringVar(&_ip, "ip", "127.0.0.1", "本机IP")
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

	// 创建默认的文章
	cmd := exec.Command(_hugo, "new", fmt.Sprintf("%s/%s.md", _post, name))
	cmd.Dir = _home
	err := cmd.Run()
	if err != nil {
		result.Code = 10001
		result.Info = fmt.Sprintf("failed with %s\n", err)
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	tmpFilePath := fmt.Sprintf("%s/%s/%s.md", _home, _post, name)
	result.Info, _ = os.ReadFile(tmpFilePath)
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

	tmpFilePath := fmt.Sprintf("%s/%s/%s.md", _home, _post, name)
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

	err = os.WriteFile(tmpFilePath, decodedBytes, os.ModePerm)
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
		files = append(files, strings.TrimRight(file.Name(), ".md"))
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

	tmpFilePath := fmt.Sprintf("%s/%s/%s.md", _home, _post, name)
	_, err := os.Stat(tmpFilePath)
	if os.IsNotExist(err) {
		result.Code = 40000
		result.Info = "posts not existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	result.Code = 200
	result.Info, _ = os.ReadFile(tmpFilePath)
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

func deleteFolder(folderPath string) error {
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.RemoveAll(path)
	})

	if err != nil {
		return err
	}

	return os.Remove(folderPath)
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

	tmpFilePath := fmt.Sprintf("%s/%s/%s.md", _home, _post, name)
	os.Remove(tmpFilePath)

	imgdir := fmt.Sprintf("%s/public/MdImg/%s", _home, name)
	deleteFolder(imgdir)

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

	// 设置博客主页地址
	hugoTomlPath := fmt.Sprintf("%s/%s", _home, "hugo.toml")
	_, err := os.Stat(hugoTomlPath)
	if os.IsNotExist(err) {
		result.Code = 60000
		result.Info = "hugo.toml not existed."
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	cfg, err := os.ReadFile(hugoTomlPath)
	if err != nil {
		result.Code = 60001
		result.Info = "hugo.toml open err:" + err.Error()
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	lines := strings.Split(fmt.Sprintf("%s", cfg), "\n")
	b := false
	index := -1
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "baseURL") {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "baseURL") {
				b = true
				index = i
			}
		}
	}

	var newLines []string = make([]string, 0)

	if !b {
		newLines = append(newLines, fmt.Sprintf("baseURL = \"http://v%s:%s/blog\"", _ip, _port))
	}

	for i := 0; i < len(lines); i++ {
		if i == index && b {
			newLines = append(newLines, fmt.Sprintf("baseURL = \"http://%s:%s/blog\"", _ip, _port))
			continue
		}
		newLines = append(newLines, fmt.Sprintf("%s", lines[i]))
	}

	err = os.WriteFile(hugoTomlPath, []byte(strings.Join(newLines, "\n")), os.ModePerm)
	if err != nil {
		result.Code = 60002
		result.Info = "hugo.toml rewrite err:" + err.Error()
		bytes, _ := json.Marshal(result)
		w.Write(bytes)
		return
	}

	cmd := exec.Command(_hugo)
	cmd.Dir = _home
	err = cmd.Run()
	if err != nil {
		result.Code = 60003
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

var extList = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".png":  true,
	".bmp":  true,
	".webp": true,
}

// 图片上传接口的特殊返回结构
type ImgBackJSON struct {
	Success int    `json:"success"`
	Message string `json:"message"`
	ImgUrl  string `json:"url"`
	ImgLink string `json:"link"`
}

func ImgBackInfo(status int, msg string, url string, link string) []byte {
	jInfo := ImgBackJSON{
		Success: status,
		Message: msg,
		ImgUrl:  url,
		ImgLink: link,
	}

	bInfo, _ := json.Marshal(jInfo)
	return bInfo
}

// 图片上传
func ImgUpdate(w http.ResponseWriter, r *http.Request) {

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

	if name == "" {
		// 文章名称为空
		w.Write(ImgBackInfo(0, "文章名称为空", "", ""))
		return
	}

	if !check(pass) {
		// 密码错误
		w.Write(ImgBackInfo(0, "密码错误", "", ""))
		return
	}

	f, h, e := r.FormFile("editormd-image-file")
	if e != nil {
		// 接收数据错误
		w.Write(ImgBackInfo(0, "接收数据错误", "", ""))
		return
	}
	defer f.Close()

	ext := strings.ToLower(path.Ext(h.Filename))
	_, bExt := extList[ext]
	if !bExt {
		// 不接受的后缀
		w.Write(ImgBackInfo(0, "不接受的后缀", "", ""))
		return
	}

	// 创建文件
	dstName := fmt.Sprintf("%s/public/MdImg/%s", _home, name)
	_, err := os.Stat(dstName)
	if os.IsNotExist(err) {
		os.MkdirAll(dstName, os.ModePerm)
	}

	dst, err := os.Create(dstName + "/" + h.Filename)
	if err != nil {
		// 创建本地文件失败
		w.Write(ImgBackInfo(0, "创建本地文件失败", "", ""))
		return
	}
	defer dst.Close()

	// 将文件保存到服务器
	_, err = io.Copy(dst, f)
	if err != nil {
		// 保存文件失败
		w.Write(ImgBackInfo(0, "保存文件失败", "", ""))
		return
	}

	url := fmt.Sprintf("/blog/MdImg/%s/%s", name, h.Filename)
	w.Write(ImgBackInfo(1, "上传成功", url, url))
}

// 重定向主页
func redirectIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pages/editormd/public", http.StatusSeeOther)
}

func main() {
	// 主页重定向
	http.HandleFunc("/editor-helper", redirectIndex)
	// 功能性接口
	http.HandleFunc("/new", MdNew)
	http.HandleFunc("/save", MdSave)
	http.HandleFunc("/list", MdList)
	http.HandleFunc("/open", MdOpen)
	http.HandleFunc("/del", MdDel)
	http.HandleFunc("/hugo", Hugo)
	http.HandleFunc("/imgupdate", ImgUpdate)
	http.HandleFunc("/ischeck", IsCheck)
	// 代理editor.md的静态资源
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.FS(editormdFS))))
	http.Handle("/blog/", http.StripPrefix("/blog/", http.FileServer(http.Dir(_home+"/public"))))

	fmt.Println("hugo管理助手已启动 端口:", _port)
	fmt.Printf("    博客主页地址 :http://%s:%s/blog\n", _ip, _port)
	fmt.Printf("    编辑主页地址 :http://%s:%s/editor-helper\n", _ip, _port)
	err := http.ListenAndServe(":"+_port, nil)
	if err != nil {
		fmt.Println("服务器开启错误: ", err)
	}
}
