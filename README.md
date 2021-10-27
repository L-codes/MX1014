# MX1014

**MX1014** 是一个遵循 **“短平快”** 原则的灵活、轻便和快速端口扫描器

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## Version

2.1.0 - [版本修改日志](CHANGELOG.md)


## Features

* 兼容 nmap 的端口和目标语法，并支持导入多个 TARGET, 灵活扫描
* 扫描过程中有自动判定主机存活是否继续扫描其主机的机制，从而加快端口探测速度
* 使用端口分组的概念，方便指定特定端口组，进行针对性扫描 (端口别名，参考下面的 "Port Group")
* 支持 TCP/UDP 的 Echo 回显数据发送 (UDP 不会返回端口状态)，便于出网探测
* 支持 TCP closed 状态显示，便于主机存活与出网探测
* 支持端口模糊测试
* 支持各组目标扫描不同的端口
* windows 最低环境支持 xp/2003 等 (即兼容 Golang 1.10.8)
* linux 支持 CentOS5 (Linux 2.6.18) 等 (即兼容 Golang 1.10.8)


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
  .001              1001               Version 2.1.0
  .1.              ...1.


Usage: ./mx1014 [Options] [Target1] [Target2]...

Target Example:
    192.168.1.0/24
    192.168.1.*
    192.168.1-12.1
    192.168.*.1:22,80-90,8080
    github.com:22,443,rce

Options:
  [Target]
    -i  File   Target input from list
    -g  Net    Intranet gateway address range (10/100/172/192/all)
    -sh        Show scan target
    -cnet      C net mode

  [Port]
    -p  Ports  Default port ranges (Default is "in" port group)
    -sp        Only show default ports (see -p)
    -ep Ports  Exclude port (see -p)
    -hp Ports  Priority scan port (Default 80,443,8080,22,445,3389)
    -fuzz      Fuzz Port

  [Connect]
    -t  Int    The Number of Goroutine (Default is 512)
    -T  Int    TCP Connect Timeout (Default is 1980ms)
    -u         UDP spray
    -e         Echo mode (TCP needs to be manually)
    -A         Disable auto discard
    -a  Int    Too many filtered, Discard the host (Default is 512)

  [Output]
    -o  File   Output file path
    -c         Allow display of closed ports (Only TCP)
    -d  Str    Specify Echo mode data (Default is "%port%\n")
    -D  Int    Progress Bar Refresh Delay (Default is 5s)
    -l         Output alive host
    -v         Verbose mode
```

2. 简单扫描三百多个内网常见端口
```ruby
$ ./mx1014 192.168.1.134
# 2021/10/16 15:30:13 Start scanning 1 hosts... (reqs: 311)

192.168.1.134:22       (in,mac,brute,info,ssh,win,linux)
192.168.1.134:8080     (in,web1,jdwp,rce,jboss,web2)
192.168.1.134:8009     (in,ajp,rce,web2)

# 2021/10/16 15:30:14 Finished 311 tasks. alive: 100% (1/1), open: 3, pps: 305, time: 1s
```

3. 扫描各组不同 IP 的不同端口
```ruby
$ ./mx1014 192.168.1.133/24:ssh 192.168.1.133:80-90,443,mysql
# 2021/10/16 15:47:00 Start scanning 257 hosts... (reqs: 527)

192.168.1.134:22       (win,linux,mac,brute,ssh,info,in)
192.168.1.133:22       (win,linux,mac,brute,ssh,info,in)
192.168.1.133:3307     (database1,mysql,brute,database2,in)
192.168.1.133:82       (web2,in)
192.168.1.133:3306     (database1,mysql,brute,database2,in)
192.168.1.133:80       (iis,web2,web1,jboss,in)
192.168.1.133:81       (web2,in)
192.168.1.133:83       (web2,in)
192.168.1.133:84       (web2,in)
192.168.1.133:87       (web2,in)
192.168.1.133:88       (win,web2,brute,kerberos,in)
192.168.1.133:443      (iis,web2,web1,in)

# 2021/10/16 15:47:02 Finished 527 tasks. alive: 1% (3/257), open: 12, pps: 345, time: 1s
```


## Advanced Usage
1. 根据网络环境，调整扫描并发数(-t)、超时(-T)和进度打印间隔(-D), 提速或提高准确度
```ruby
$ ./mx1014 -t 1000 -T 500 -D 10 -p 1-65535 192.168.1.134
# 2021/10/16 15:49:16 Start scanning 1 hosts... (reqs: 65535)

192.168.1.134:8009     (rce,in,web2,ajp)
192.168.1.134:22       (mac,ssh,brute,in,linux,win,info)
192.168.1.134:8080     (jboss,jdwp,rce,in,web2,web1)
# Progress (19022/65535) open: 3, pps: 1895, rate: 29% (RD 24s)
# Progress (38701/65535) open: 3, pps: 1931, rate: 59% (RD 13s)
# Progress (58673/65535) open: 3, pps: 1953, rate: 90% (RD 3s)

# 2021/10/16 15:49:50 Finished 65535 tasks. alive: 100% (1/1), open: 3, pps: 1951, time: 33s
```

2. TCP Echo 模式，如果端口开放，往端口写入当前的端口号
```ruby
$ ./mx1014 -e 192.168.1.134:80,8000-8080 # 可用 -d 参数指定 echo 的内容
# 2021/10/16 15:52:58 Start scanning 1 hosts... (TCP Echo) (reqs: 82)

192.168.1.134:8080     (web2,jboss,jdwp,rce,in,web1)
192.168.1.134:8009     (web2,rce,in,ajp)

# 2021/10/16 15:52:59 Finished 82 tasks. alive: 100% (1/1), open: 2, pps: 81, time: 1s
```

3. 从文件中读取目标并进行 UDP 扫描 (默认会 echo 端口号; 可用于出网端口测试)
> VPS 可利用下面的转发方便接受 echo 内容
> `iptables -t nat -A PREROUTING -p udp -m multiport --dports 80,8000:8080 -j REDIRECT --to-port 666`
```ruby
$ cat > ip.txt <<EOF
heredoc> 192.168.1.134:80
heredoc> 192.168.1.130:22
EOF
$ ./mx1014 -u -i ip.txt
# 2021/10/16 15:57:47 Start scanning 2 hosts... (UDP Spray) (reqs: 2)


# 2021/10/16 15:57:47 Finished 2 tasks. alive: 0% (0/2), open: 0, pps: 1306, time: 0s
```

4. 可显示 closed 状态的端口信息，作用自行脑补
```ruby
$ ./mx1014 -c 192.168.1.134:1-65535
```

5. 禁用自动丢弃主机机制，强制扫描
```ruby
$ ./mx1014 -A 192.168.1.134:1-65535
```

6. 快速探测内网资产
```ruby
# 通过 80 端口找到内网存活的网段
$ ./mx1014 -l -p 80 -g all -o up.txt
# 根据存活的网段进行 C 段探测
$ ./mx1014 -cnet -i up.txt
```

7. 生成模糊的相近端口进行扫描
```ruby
# 可根据端口生成相近可能的端口
$ ./mx1014 -sp -p 80 -fuzz
# Count: 4
81,80,8080,79
```


## Port Group
```ruby
# NOTE Reference:
#   https://book.hacktricks.xyz/pentesting/
#   https://github.com/0xtz/Enum_For_All
{
  # pentest
  in: "rce,info,brute,web2",
  rce: "rlogin,jndi,nfs,oracle_ftp,docker,squid,cisco,glassfish,altassian,hp,vnc,nodejs_debug,redis,jdwp,ajp,zabbix,nexus,activemq,zoho,hashicorp,solr,php_xdebug,kafka,elasticsearch,vmware,rocketmq,lpd,distcc,epmd,ipmi,modbus,smb",
  info: "ftp,ssh,telnet,mail,snmp,rsync,lotus,zookeeper,kibana,pcanywhere,hadoop,checkpoint,iscsi,saprouter,svn,rpc,rusersd,rtsp,amqp,msrpc,netbios",
  brute: "ftp,ssh,smb,winrm,rsync,vnc,redis,rdp,database1,telnet,mail,rtsp,kerberos,ldap,socks",

  # web
  web1: "80,443,8080",
  web2: "81-90,444,800,801,1024,2000,2001,3001,4430,4433,4443,5000,5001,5555,5800,6000-6003,6080,6443,6588,6666,7004-7009,7080,7443,7777,8000-8030,8040,8060,8066,8070,8080-8111,8181,8182,8200,8282,8363,8761,8787,8800,8848,8866,8873,8881-8890,8899,8900,8989,8999,9000-9010,9999,10000,10001,10080,10800,18080,activemq,arl,baota,cassini,dlink,ejinshan,fastcgi,flink,fortigate,hivision,ifw8,iis,java_ws,jboss,kc_aom,kibana,natshell,nexus,oracle_web,portainer,rabbitmq,rizhiyi,sapido,seeyon,solr,squid,weblogic,websphere_web,yapi,elasticsearch,zabbix",
  iis: "80,443,47001",
  jboss: "80,1111,4444,4445,8080,8443,45566",
  zookeeper: "2181,2888,3888",
  solr: "8983",
  websphere_web: "8880,9043,9080,9081,9082,9083,9090.9091,9443",
  websphere: "websphere_web,2809,5558,5578,7276,7286,9060,9100,9353,9401,9402",
  activemq: "8161",
  weblogic: "7000,7001,7002,7003,7010,7070,7071",
  squid: "3128",
  rabbitmq: "15672",
  flink: "8081",
  oracle_web: "3339",
  baota: "888,8888",
  fastcgi: "9000",
  kc_aom: "12580",
  kibana: "5601",
  portainer: "9000",
  natshell: "7788",
  elasticsearch: "9200,9300",
  rizhiyi: "8180",
  arl: "5003",
  cassini: "6868",
  dlink: "55555",
  fortigate: "10443",
  nexus: "8081",
  sapido: "1080",
  yapi: "3000",
  hivision: "7088",
  ejinshan: "6868",
  seeyon: "8001",
  java_ws: "8887",
  ifw8: "880",
  zabbix: "8069",

  # mail
  mail: "smtp,pop2,pop3,imap",
  pop2: "109",
  pop3: "110,995",
  imap: "143,993",
  smtp: "25,465,587,2525",

  # database
  database1: "mssql,oracle,mysql,postgresql,redis,memcache,mongodb",
  database2: "mssql,oracle,mysql,sybase,db2,postgresql,couchdb,redis,memcache,hbase,mongodb,hsqldb,cassandra",
  mysql: "3306,3307,3308",
  mssql: "1433,1434",
  oracle: "210,1158,1521",
  hsqldb: "9001",
  redis: "6379,63790",
  postgresql: "5432",
  mongodb: "27017,28017",
  db2: "5000",
  sybase: "4100,5000",
  couchdb: "5984",
  memcache: "11211",
  hbase: "16000,16010,16020,16030",
  cassandra: "9042,9160",

  # os
  win: "ssh,ftp,telnet,kerberos,msrpc,vnc,netbios,ldap,smb,socks,rdp,winrm,ntp",
  linux: "ssh,ftp,telnet,rlogin,vnc,x11,nfs,whois,socks,ntp,isakmp,rsync,rpc,ipmi,rusersd",
  mac: "ssh,afp,vnc,nfs",

  # other
  kerberos: "88",
  netbios: "137,138,139",
  smb: "139,445",
  rdp: "3389",
  winrm: "5985,5986",
  afp: "548",
  ftp: "21,115,2121",
  whois: "43",
  dns: "53",
  socks: "1080",
  oracle_ftp: "2100",
  ssh: "22,2222",
  ntp: "123",
  isakmp: "500",
  printer: "9100",
  mqtt: "1883",
  ajp: "8009",
  vnc: "5800,5900,5901",
  rsync: "873",
  nfs: "2049",
  sangfor: "51111",
  nodejs_debug: "5858,9229",
  telnet: "23",
  rpc: "111",
  msrpc: "135,593",
  irc: "194,6660",
  ldap: "389,636,3268,3269",
  modbus: "502",
  rtsp: "554,8554",
  ipmi: "623",
  rusersd: "1026",
  amqp: "5672",
  kafka: "9092",
  hp: "5555,5556",
  altassian: "4990",
  lotus: "1352",
  cisco: "4786",
  lpd: "515",
  php_xdebug: "9000",
  hashicorp: "8500",
  checkpoint: "264",
  pcanywhere: "5632",
  docker: "2375,2376,2377,5000",
  iscsi: "3260",
  saprouter: "3299",
  distcc: "3632",
  zoho: "8383",
  svn: "3690",
  snmp: "161",
  epmd: "4369",
  hadoop: "8020,8040,8041,8042,8480,8485,9000,9083,19888,41414,50010,50020,50070,50075,50090,50470,50475",
  rmi: "1028,1098,1090,4444,11099,47001,10999,1099",
  jndi: "rmi,1000,1001,1100,1101,4444,4445,4446,4447,5001,8083,9999,10001,10999,11099,19001",
  jmx: "8093,8686,9010,9011,9012,50500,61616",
  jdwp: "5005,8000,8080,8453,45000,45001",
  rlogin: "512,513,514",
  glassfish: "4848",
  rocketmq: "9876,10909,10911,10912",
  vmware: "9875",
  x11: "6000",
}
```

## TODO

 * 代码逻辑优化

 * 实现 SYN ACK NULL 等 raw socket 扫描 (寻找好的实现方案中)

 * 继续优化端口组列表，望大家共同维护


## License

GPL 3.0
