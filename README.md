# MX1014

**MX1014** 是一个遵循 **“短平快”** 原则的灵活、轻便和快速端口扫描器

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## Version

1.1.1 - [版本修改日志](CHANGELOG.md)



## Features

* 兼容 nmap 的端口和目标语法
* 支持各组目标扫描不同的端口
* 采用逐主机的深度搜索机制，降低 “踩雷” 的速度
* 扫描过程中有自动判定主机存活是否继续扫描的机制
* 支持导入多个 TARGET
* 默认主机和端口均为随机循序扫描
* windows 最低环境支持 xp/2003 等 (即兼容 Golang 1.10.8)
* linux 支持 CentOS5 (Linux 2.6.18) 等 (即兼容 Golang 1.10.8)
* 支持 TCP/UDP 的 Echo 回显数据发送 (UDP 不会返回端口状态)
* 支持 TCP closed 状态显示



## Basic Usage
1. 直接运行，查看帮助信息 (所有参数与语法说明)
```ruby
$ ./mx1014
                          ...                                     .
                        .111111111111111.........................1111
      ......111..    .10011111011111110000000000000000111111111100000
  10010000000011.1110000001.111.111......1111111111111111..........
  10twelve0111...   .10001. ..
  100011...          1001               MX1014 by L
  .001              1001               Version 1.1.1
  .1.              ...1.


Usage: ./mx1014 [Options] [Target1] [Target2]...

Target Example:
    192.168.1.0/24
    192.168.1.*
    192.168.1-12.1
    192.168.*.1:22,80-90,8080
    github.com:22,443,8443

Options:
    -p Ports  Default port ranges. (Default is common ports
    -ap Ports Append default ports
    -i File   Target input from list
    -t Int    The Number of Goroutine (Default is 256)
    -T Int    TCP Connect Timeout (Default is 1014ms)
    -o File   Output file path
    -r        Scan in import order
    -u        UDP spray
    -e        Echo mode (TCP needs to be manually)
    -c        Allow display of closed ports (Only TCP)
    -d Str    Specify Echo mode data (Default is "%port%\n")
    -D Int    Progress Bar Refresh Delay (Default is 5s)
    -a Int    Too many filtered, Discard the host (Default is 1014)
    -A        Disable auto disable
    -v        Verbose mode
```

2. 简单扫描数十个常用默认端口
```ruby
$ ./mx1014 192.168.1.134
# 2021/04/09 12:15:49 Start scanning 1 hosts...

192.168.1.134:80
192.168.1.134:8009
192.168.1.134:22
192.168.1.134:8080

# 2021/04/09 12:15:50 Finished 49 tasks. alive: 100% (1/1), open: 4, pps: 49, time: 1s
```

3. 扫描各组不同 IP 的不同端口
```ruby
$ ./mx1014 192.168.1.0/24:22 192.168.1.133:80-90,443
# 2021/04/09 12:20:57 Start scanning 257 hosts...

192.168.1.133:83
192.168.1.133:84
192.168.1.133:81
192.168.1.133:22
192.168.1.133:443
192.168.1.134:22
192.168.1.133:80
192.168.1.133:87
192.168.1.133:82
192.168.1.133:88
192.168.1.130:22

# 2021/04/09 12:20:58 Finished 268 tasks. alive: 1% (4/257), open: 11, pps: 263, time: 1s
```


## Advanced Usage
1. 根据网络环境，调整扫描并发数(-t)、超时(-T)和进度打印间隔(-D), 提速或提高准确度
```ruby
$ ./mx1014 -t 1000 -T 500 -D 10 -p 1-65535 192.168.1.134
# 2021/04/07 19:31:40 Start scanning 1 hosts...

192.168.1.134:3306
192.168.1.134:80
# Progress (19021/65535) open: 2, pps: 1900, rate: 29%
192.168.1.134:8009
# Progress (39042/65535) open: 3, pps: 1951, rate: 59%
192.168.1.134:22
# Progress (58270/65535) open: 4, pps: 1941, rate: 88%
192.168.1.134:8080

# 2021/04/07 19:32:13 Finished 65535 tasks. alive: 100% (1/1), open: 5, pps: 1934, time: 33s
```

2. TCP Echo 模式，如果端口开放，往端口写入当前的端口号
```ruby
$ ./mx1014 -e 192.168.1.134:80,8000-8080 # 可用 -d 参数指定 echo 的内容
# 2021/04/07 19:37:43 Start scanning 1 hosts... (TCP Echo)

192.168.1.134:8009
192.168.1.134:8011
192.168.1.134:8080
192.168.1.134:80

# 2021/04/07 19:37:44 Finished 82 tasks. alive: 100% (1/1), open: 4, pps: 81, time: 1s
```

3. 从文件中读取目标并进行 UDP 扫描 (默认会 echo 端口号)
```ruby
$ cat > ip.txt <<EOF
heredoc> 192.168.1.134:80
heredoc> 192.168.1.130:22
EOF
$ ./mx1014 -u -i ip.txt
# 2021/04/07 19:50:39 Start scanning 2 hosts...


# 2021/04/07 19:50:39 Finished 2 tasks. alive: 0% (0/2), open: 0, pps: 1306, time: 0s
```

4. 追加默认端口进行扫描
```ruby
$ ./mx1014 -ap 1000-2000 192.168.1.134
# 2021/04/09 12:34:27 Start scanning 1 hosts...

192.168.1.134:8009
192.168.1.134:80
192.168.1.134:8080
192.168.1.134:22
# Progress (1032/1050) open: 4, pps: 206, rate: 98%
192.168.1.134:1111

# 2021/04/09 12:34:32 Finished 1050 tasks. alive: 100% (1/1), open: 4, pps: 208, time: 5s
```

5. 禁用自动丢弃主机机制，强制扫描
```ruby
$ ./mx1014 -A 192.168.1.134:1-65535
```

6. 可显示 closed 状态的端口信息，作用自行脑补
```ruby
$ ./mx1014 -c 192.168.1.134:1-65535
```


## TODO

 * 代码逻辑优化

 * 增加 nmap -PT 等存活探测功能


## License

GPL 3.0
