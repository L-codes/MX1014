# MX1014

**MX1014** 是一个遵循 **“短平快”** 原则的灵活、轻便和快速端口扫描器

> 此工具仅限于安全研究和教学，用户承担因使用此工具而导致的所有法律和相关责任！ 作者不承担任何法律和相关责任！


## Version

1.2.0 - [版本修改日志](CHANGELOG.md)



## Features

* 兼容 nmap 的端口和目标语法
* 支持各组目标扫描不同的端口
* 支持端口组进行针对性的扫描
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
  .001              1001               Version 1.2.0
  .1.              ...1.


Usage: ./mx1014 [Options] [Target1] [Target2]...

Target Example:
    192.168.1.0/24
    192.168.1.*
    192.168.1-12.1
    192.168.*.1:22,80-90,8080
    github.com:22,443,8443

Options:
    -p  Ports  Default port ranges. (Default is common ports)
    -ap Ports  Append default ports
    -i  File   Target input from list
    -t  Int    The Number of Goroutine (Default is 256)
    -T  Int    TCP Connect Timeout (Default is 1514ms)
    -o  File   Output file path
    -r         Scan in import order
    -u         UDP spray
    -e         Echo mode (TCP needs to be manually)
    -c         Allow display of closed ports (Only TCP)
    -d  Str    Specify Echo mode data (Default is "%port%\n")
    -D  Int    Progress Bar Refresh Delay (Default is 5s)
    -a  Int    Too many filtered, Discard the host (Default is 1014)
    -A         Disable auto disable
    -v         Verbose mode
    -sp        Only show default ports
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

6. 根据 TCP 探测存活主机
```ruby
$ ./mx1014 -l -p 80 192.168.1.134
```

## Port Group
```ruby
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


## License

GPL 3.0
