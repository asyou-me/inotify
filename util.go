package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//从.inotify文件 获取 主文件夹下的指定的文件夹 的所有子文件夹
func GetDirs(path string) (*[]string, error) {
	dir := make([]string, 0, 30)
	var err error

	//没有路径 默认当前目录
	if path == "" {
		path, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return &dir, err
		}
	}

	fi, err := os.Open(path + "/.inotify")

	if err == nil {
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

//转换成绝对路径并验证文件是否存在
func file_path_check(path *string) error {
	isrelative := strings.HasSuffix(*path, "/")
	if !isrelative {
		curr_path, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		*path = curr_path + "/" + *path
	}

	if !Exist(*path) {
		return errors.New("目标shell文件不存在")
	}
	return nil
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
