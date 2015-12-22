package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetDirs(path string) (*[]string, error) {
	dir := make([]string, 0, 30)
	var err error

	if path == "" {
		path, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return &dir, err
		}
	}

	fi, err := os.Open(path + "/.inotify")

	config_dirs := make([]string, 0, 30)
	if err != nil {
		return &dir, err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		return &dir, err
	}

	if err = json.Unmarshal(fd, &config_dirs); err != nil {
		if err != nil {
			return &dir, err
		}
	}
	for _, v := range config_dirs {
		WalkDir(path+"/"+v, &dir)
	}

	dir = append(dir, path)

	return &dir, nil
}

//获取指定目录及所有子目录。
func WalkDir(dirPth string, dirs *[]string) (err error) { //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if fi.IsDir() {
			*dirs = append(*dirs, filename)
		}
		return err
	})
	return err
}
