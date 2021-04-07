# MX1014

**MX1014** 是一个灵活、轻便和快速端口扫描器

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## Version

1.0.0 - [版本修改日志](CHANGELOG.md)



## Features

* 兼容 nmap 的端口和目标语法
* 支持各组目标扫描不同的端口
* 支持导入多个 TARGET
* 支持 TCP/UDP 的 Echo 回显数据发送 (UDP 不会返回端口状态)
* windows 最低环境支持 xp/2003 等 (即兼容 Golang 1.10.8)
* 默认主机和端口均为随机循序扫描



## Basic Usage
1. 直接运行，查看帮助信息 (所有参数的与语法说明)
```ruby
$ ./mx1014

                          ...                                     .
                        .111111111111111.........................1111
      ......111..    .10011111011111110000000000000000111111111100000
  10010000000011.1110000001.111.111......1111111111111111..........
  10twelve0111...   .10001. ..
  100011...          1001               MX1014 by L
  .001              1001               Version 1.0.0
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
    -i File   Target input from list
    -t Int    The Number of Goroutine (Default is 256)
    -T Int    TCP Connect Timeout (Default is 1014ms)
    -o File   Output file path
    -r        Scan in import order
    -u        UDP spray
    -e        Echo mode (TCP needs to be manually)
    -d Str    Specify Echo mode data (Default is "%port%\n")
    -D Int    Progress Bar Refresh Delay (Default is 5s)
    -v        Verbose mode
```

2. 简单扫描数十个常用默认端口
```ruby
$ ./mx1014 172.16.178.134
# 2021/04/07 18:38:45 Start scanning 1 hosts...

172.16.178.134:8080
172.16.178.134:8009
172.16.178.134:80
172.16.178.134:22

# Finished. host: 1, task: 49, open: 4, pps: 49, time: 1.01s
```

3. 扫描各组不同 IP 的不同端口
```ruby
$ ./mx1014 172.16.178.0/24:22 172.16.178.133:80-90,443
# 2021/04/07 18:41:45 Start scanning 257 hosts...

172.16.178.133:82
172.16.178.133:88
172.16.178.133:81
172.16.178.133:83
172.16.178.133:84
172.16.178.133:443
172.16.178.133:22
172.16.178.134:22
172.16.178.133:80
172.16.178.133:87

# Finished. host: 257, task: 13, open: 10, pps: 13, time: 1.02s
```


## Advanced Usage
1. 根据网络环境，调整扫描并发数(-t)、超时(-T)和进度打印间隔(-D), 提速或提高准确度
```ruby
$ ./mx1014 -t 1000 -T 500 -D 10 -p 1-65535 172.16.178.134
# 2021/04/07 19:31:40 Start scanning 1 hosts...

172.16.178.134:3306
172.16.178.134:80
# Progress (19021/65535) open: 2, pps: 1900, rate: 29%
172.16.178.134:8009
# Progress (39042/65535) open: 3, pps: 1951, rate: 59%
172.16.178.134:22
# Progress (58270/65535) open: 4, pps: 1941, rate: 88%
172.16.178.134:8080

# Finished. host: 1, task: 65535, open: 5, pps: 1934, time: 33.88s
```

2. TCP Echo 模式，如果端口开放，往端口写入当前的端口号
```ruby
$ ./mx1014 -e 172.16.178.134:80,8000-8080 # 可用 -d 参数指定 echo 的内容
# 2021/04/07 19:37:43 Start scanning 1 hosts... (TCP Echo)

172.16.178.134:8009
172.16.178.134:8011
172.16.178.134:8080
172.16.178.134:80

# Finished. host: 1, task: 82, open: 4, pps: 81, time: 1.01s
```

3. 从文件中读取目标并进行 UDP 扫描 (默认会 echo 端口号)
```ruby
$ cat > ip.txt <<EOF
heredoc> 172.16.178.134:80
heredoc> 172.16.178.130:22
EOF
$ ./mx1014 -u -i ip.txt                                                                                                                                                                   # 2021/04/07 19:50:39 Start scanning 2 hosts...


# Finished. host: 2, task: 2, open: 0, pps: 1556, time: 0.00s
```


## TODO

 * 对主机先进行存活探测，再进行端口扫描


## License

GPL 3.0
