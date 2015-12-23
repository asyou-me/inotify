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

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	path := flag.String("path", "", "path like /tmp")
	shell := flag.String("shell", "", "a shell file")
	h := flag.Bool("h", false, "help")

	flag.Parse()

	if *h == true {
		Usage()
		os.Exit(0)
	}

	*path = strings.TrimRight(*path, "/")

	dirs, err := GetDirs(*path)
	if err != nil {
		Usage()
		os.Exit(0)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Usage()
		os.Exit(0)
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
			os.Exit(0)
		}
	}

	s := <-c
	fmt.Println(" exit with", s)
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
		fmt.Println("\033[31;1mStart for Cmd", err)
	} else {
		cmd_chan <- cmd
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("\033[31;1mWait for Cmd", err)
	}
}
