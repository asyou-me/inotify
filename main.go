package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

var cmd_chan = make(chan *exec.Cmd)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

//实现对指定文件夹的监听，如果有修改或变动重新编译并执行
func main() {
	//设置一个channel来发送信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	// 一直运行一直到收到一个信号

	//分析命令行的参数（被监听文件夹名  脚本文件名）
	path := flag.String("path", "", "path like /tmp")
	shell := flag.String("shell", "", "a shell file")
	h := flag.Bool("h", false, "help")

	flag.Parse()

	if *h == true {
		Usage()
		os.Exit(0)
	}

	//获取要监听的文件夹名
	*path = strings.TrimRight(*path, "/")

	//从.inotify文件 获取 要监听的文件夹下的指定文件夹 的所有子文件夹
	dirs, err := GetDirs(*path)
	if err != nil {
		fmt.Println(err)
		Usage()
		os.Exit(0)
	}

	err = file_path_check(shell)
	if err != nil {
		Usage()
		os.Exit(0)
	}

	//建立一个监听者
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Usage()
		os.Exit(0)
	}
	defer watcher.Close()

	//独立进程来监听和执行编译脚本
	go func() {
		var last_time int64 = time.Now().Unix()
		var run_num int64 = 0
		var run_cmd *exec.Cmd
		go run(shell)

		//循环监听修改次数
		for {
			select {
			case _ = <-watcher.Events:
				run_num++
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			case run_cmd = <-cmd_chan:

			}

			//5秒或5秒内修改超过10次，重启进程
			if run_num > 10 || (run_num > 0 && time.Now().Unix() > last_time+5) {
				run_cmd.Process.Kill()
				run_num = 0
				last_time = time.Now().Unix()
				go run(shell)
			}
		}
	}()

	//添加 指定监听文件夹下的 所有子文件夹 到监听者
	for _, v := range *dirs {
		err = watcher.Add(v)
		if err != nil {
			Usage()
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
		cmd_chan <- cmd //把执行命令的进程名放进管道 以便停止该进程
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("\033[31;1mWait for Cmd", err)
	}
}
