# inotify

golang开发，需要不断的重复固定的流程（修改代码->编译代码），执行go build xxx ./xxx会耗费很多的开发时间。inotify是一个自动编译golang源代码的工具。linux/mac上都支持shell，通过inotify监听项目文件的变化，然后inotify调用shell脚本编译golang执行

### Version
0.01

### Tech

inotify使用了如下的开源项目:

* [golang] - github.com/golang/go
* [fsnotify] - gopkg.in/fsnotify.v1

inotify源代码被开放在GitHub

### Installation

二进制安装（linux）:

 - ```sh
    # 获取编译后的二进制可执行文件（linux）
    $ wget https://github.com/asyoume/inotify/releases/download/0.01/inotify
    # 将获取到的二进制文件放到一个可用的PATH目录（例如/usr/bin/inotify）
    $ mv inotify /xxx/xxx
    ```

源码编译（mac/windows/linux）：

 - 1.安装golang编译环境（http://golang.org/doc/install）
 - 2.获取项目依赖
    ```sh
    $ go get gopkg.in/fsnotify.v1
    ```
 - 3.获取项目的源码
    ```sh
    $ go get github.com/asyoume/inotify
    ```
 - 4.编译项目源码
    ```sh
    $ go install github.com/asyoume/inotify
    ```
 
### use

 - inotify的参数
   - -shell （shell脚本的路径）
   - -path （项目所在的目录）
 - inotify项目所在目录需要创建一个名为 .inotify 的文件，指定需要监控的目录。例如：
     ```javascript
    [
      "source"
    ]
    ```
 - 例子
    ```sh
    $ cd "$GOPATH/src/github.com/asyoume/inotify/example"
    $ inotify -shell shell/run.sh
    ```

### Todos

 - 使用更少的cpu
 - 暂时有进程sleep的状况

License