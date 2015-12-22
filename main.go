package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/exec"
	"strings"
	"time"
)

var done = make(chan bool)
var cmd_chan = make(chan *exec.Cmd)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	path := flag.String("path", "", "path like /tmp")
	shell := flag.String("shell", "", "a shell file")

	flag.Parse()

	*path = strings.TrimRight(*path, "/")

	dirs, err := GetDirs(*path)
	if err != nil {
		Usage()
		panic(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Usage()
		panic(err)
	}
	defer watcher.Close()

	go func() {
		var last_time int64 = time.Now().Unix()
		var run_num int64 = 0
		var run_cmd *exec.Cmd
		go run(shell)
		for {
			select {
			case _ = <-watcher.Events:
				run_num++
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			case run_cmd = <-cmd_chan:

			}
			if run_num > 10 || (run_num > 0 && time.Now().Unix() > last_time+5) {
				run_cmd.Process.Kill()
				run_num = 0
				last_time = time.Now().Unix()
				go run(shell)
			}
		}
	}()

	for _, v := range *dirs {
		err = watcher.Add(v)
		if err != nil {
			Usage()
			panic(err)
		}
	}

	<-done
}

func run(shell *string) {
	cmd := exec.Command("sh", *shell)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("\033[31;1m Error creating StdoutPipe for Cmd", err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Println("\033[31;1mError Start for Cmd", err)
	} else {
		cmd_chan <- cmd
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("\033[31;1mError Wait for Cmd", err)
	}
}
