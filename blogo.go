package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Comment struct {
	Title  string
	Body   []byte
	Date   string
	Author string
	Email  string
}

type Post struct {
	Title    string
	Body     []byte
	Date     string
	Comments []Comment
}

func (p *Post) create() error {
	return ioutil.WriteFile("content/posts/"+strings.Replace(p.Title, " ", "-", -1)+".txt", p.Body, 0600)
}

type byDate []os.FileInfo

func (f byDate) Len() int           { return len(f) }
func (f byDate) Less(i, j int) bool { return time.Since(f[i].ModTime()) > time.Since(f[j].ModTime()) }
func (f byDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(byDate(list))
	return list, nil
}

func readLines(path string) (lines []string, err error) {
	var (
		file   *os.File
		part   []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 1024))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func getConfigValue(key string) (value string) {
	lines, err := readLines("blog.conf")
	if err != nil {
		fmt.Println("Error: %s\n", err)
		return
	}
	for _, line := range lines {
		configData := strings.Split(line, ":")
		if len(configData) > 2 {
			fmt.Println("Check ConfigFile, and option has more than value")
			return
		}
		if configData[0] == key {
			value = configData[1]
			break
		}
	}
	return value
}

var wwwroot string = getConfigValue(("wwwroot"))

const (
	viewLen = len("/view/")
)

func getPost(title string) (post Post, error error) {
	post.Title = strings.Replace(title, "-", " ", -1)
	filename := title + ".txt"
	post.Body, error = ioutil.ReadFile("content/posts/" + filename)
	commentList, error := ReadDir("content/comments/" + title + "/")
	//var tempc []Comment
	for _, comment := range commentList {
		var c Comment
		c.Title = strings.Replace(strings.Replace(comment.Name(), "-", " ", -1), ".txt", " ", -1)
		c.Body, error = ioutil.ReadFile("content/comments/" + title + "/" + comment.Name())
		post.Comments = append(post.Comments, c)
	}
	return post, nil
}

func view(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[viewLen:]
	post, error := getPost(title)
	if error != nil {
		fmt.Println("Error: %s\n", error)
		return
	}
	t, err := template.ParseFiles("content/view.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(post.Body)
	t.Execute(w, post)
}

func new(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "content/new.html")
}

func create(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")
	p := &Post{Title: title, Body: []byte(body)}
	p.create()
	http.Redirect(w, r, "/view/"+strings.Replace(title, " ", "-", -1), http.StatusFound)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getConfigValue("wwwroot"))
	fmt.Fprintf(w, r.RequestURI)
}

func main() {
	fmt.Printf("BloGo Starting up..")
	http.HandleFunc("/view/", view)
	http.HandleFunc("/new/", new)
	http.HandleFunc("/create/", create)
	http.HandleFunc("/hello/", index_handler)
	http.ListenAndServe(":8080", nil)
}
