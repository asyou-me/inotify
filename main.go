package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"gopkg.in/fsnotify.v1"
)

var cmdChan = make(chan *exec.Cmd)

func usage() {
	fmt.Fprint(os.Stderr, "usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

//实现对指定文件夹的监听，如果有修改或变动重新编译并执行
func main() {
	//设置一个channel来发送信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	//分析命令行的参数（被监听文件夹名  脚本文件名）
	path := flag.String("path", "", "path like /tmp")
	shell := flag.String("shell", "", "a shell file")
	h := flag.Bool("h", false, "help")

	flag.Parse()

	if *h == true {
		usage()
		os.Exit(0)
	}

	//获取要监听的文件夹名
	*path = strings.TrimRight(*path, "/")

	//从.inotify文件 获取 要监听的文件夹下的指定文件夹 的所有子文件夹
	dirs, err := GetDirs(*path)
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(0)
	}

	err = file_path_check(*shell)
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(0)
	}

	//建立一个监听者
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(0)
	}
	defer watcher.Close()

	//独立进程来监听和执行编译脚本
	go func() {
		var lastTime = time.Now().Unix()
		var runCmd *exec.Cmd
		go run(shell)

		//循环监听修改次数
		for {
			select {
			// 新的事件
			case _ = <-watcher.Events:
				lastTime = time.Now().Unix()
				go func() {
					time.Sleep(3 * time.Second)
					if time.Now().Unix()-3 >= lastTime {
						lastTime = time.Now().Unix()
						if runCmd != nil {
							runCmd.Process.Kill()
							runCmd = nil
						}
						run(shell)
					}
				}()
			// 错误的事件直接忽略
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			// 执行后回掉 cmd
			case cmd := <-cmdChan:
				lastTime = time.Now().Unix()
				runCmd = cmd
			}
		}
	}()

	//添加 指定监听文件夹下的 所有子文件夹 到监听者
	for _, v := range *dirs {
		err = watcher.Add(v)
		if err != nil {
			fmt.Println(err)
			usage()
			os.Exit(0)
		}
	}

	//当停止运行时 输出
	s := <-c
	fmt.Println(" exit with", s)
}

//运行.sh脚本 （执行编译和执行）
func run(shell *string) {
	cmd := exec.Command("sh", *shell)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("\033[31;1m Error creating StdoutPipe for Cmd", err)
	}

	//转发sh的输出结果
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	//执行
	err = cmd.Start()
	if err != nil {
		fmt.Println("\033[31;1mStart for Cmd", err)
	} else {
		cmdChan <- cmd
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("\033[31;1mWait for Cmd", err)
	}
}
