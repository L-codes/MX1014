# Change Log

### v2.4.2:
    增强:
        1. 新增工控协议端口

### v2.4.1:
    增强：
        1. 增加 nacos 的端口信息
        2. 优化网络错误信息情况
        3. 添加 docker k8s 工控 等可扫描端口

### v2.4.0:
    更新简述: 更加智能方便!!!
    新特征:
        1. 新增 -r 参数，开启自动检测排除端口全开放的目标(如 synproxy)，避免无意结果出现
        2. Unix 环境运行时可自动尝试调节`ulimit -n`限制
    增强：
        1. 错误信息优化
        2. 遇到 `too many open files` 错误则结束扫描并提醒，避免无意义扫描
        3. 扫描端口列表更新
    特别感谢 @M09Ic 对 golang 低版本编译等问题的支持!

### v2.3.1:
    更新简述: 支持更多平台运行!!!
    增强:
        1. release 支持 centos5 linux 2.6.18 版本
        2. 调整 -D 默认值为 7 秒
        3. 更新了端口列表
    修复:
        1. -hp 和 -fuzz 同时使用时，-hp 指定端口不准确问题

### v2.3.0:
    更新简述: 更快，更好联动!!!
    新特征:
        1. 添加 -P 参数，不输出 open 端口的协议预判信息
    增强:
        1. 检测地址有效性，采用了并发的形式，大大提高了检测地址的速度
    修复:
        1. 优化了 help 信息
        2. 修复 target 太长导致和协议预判信息没有间隔的问题

### v2.2.0:
    新特征:
        1. 添加 -I 参数，忽略错误地址继续扫描
    增强:
        1. fuzz 功能支持左右位重叠的端口
    修复:
        1. 移除了 -g 的 100.64 网段

### v2.1.0:
    更新简述: 内网探测，更快，更准，更方便!!!

    新特征:
        1. 新增 -g 参数，可方便指定内网网关地址范围作为目标
        2. 新增 -cnet 参数，可将输入的目标地址转成 CIDR mask 24 进行扫描
        3. 新增 -hp 参数，可在随机端口扫描下指定优先扫描的端口列表
        4. 新增 -ep 参数，排除端口
        5. 新增 -sh 参数，打印扫描主机列表
    增强:
        1. 增强错误主机地址匹配，减少错误信息提示
        2. CIDR 格式的目标地址，忽略网络地址和广播地址
        3. 检测设置端口范围有效性
        4. 增加结果保存信息提醒
        5. 增强 -i 参数的目标文件列表读取，可使用 "#" 开头注释，并且目标列表从覆盖改成了追加
        6. 对可选参数进行分类，优化了 help 打印信息
    修复:
        1. 发包任务计数错误显示
        2. 修复 pps 显示 BUG

### v2.0.0:
    更新简述: 更快，更准，更方便!!!

    新特征:
        1. 引入了端口组的概念，支持 -p rce,info 等端口组进行扫描 (更多参考 README)
        2. 添加了端口模糊测试功能 (参考 -fuzz)
    增强:
        1. 开放端口的结果追加了归宿端口组的信息
        2. 统一打印输出到 stdout
        3. 调整了默认扫描并发数，由 256 提升为 512
        4. 调整了默认自动跳过主机最大 filtered 计数，由 1014 降低为 512
        4. 调整了默认Timeout，由 1514ms  提升为 1980ms
        5. -sp 打印默认端口功能，目前会根据 -fuzz -p 等选项进行输出实际扫描的默认端口
        6. 扫描前对目标地址进行校验
        7. 增强输出信息提醒
        8. 大型扫描时内存占用过大，经优化降低了 4 倍左右的内存使用
        9. 修改了原来的扫描方式(深度转为广度)，降低扫描的漏报情况
    移除:
        1. 因使用了端口组，而无需追加默认端口功能，故移除 -ap 参数
        2. 改变了新扫描方式后，只能随机端口扫描，故移除 -r 参数

### v1.2.0:
    新特征:
        1. 增加 `-l` 参数，存活主机探测模式，仅打印存活主机(不去重)
    增强:
        1. 增强扫描时间统计的显示方式
        2. 增强 `-p` 端口 - 范围解析，并校验有效值
        3. 调整了默认 timeout 时间为 1524 毫秒
        4. 增加了默认扫描端口
    修复:
        1. 修复 `-p` 时，不会优先扫描常用端口的问题

### v1.1.2:
    新特征:
        1. 增加 `-sp` 参数，打印默认扫描端口列表
    修复:
        1. 修复 help 信息内容

### v1.1.1
    新特征:
        1. 扫描进度信息中追加了估算剩余时间信息
    增强:
        1. 增加默认端口: 1024,6868,8182,9080,9999
    修复:
        1. 修复使用 `-p xx` 时，不会覆盖默认端口的逻辑问题

### v1.1.0
    概述:
        主要为自动消除 TCP 无意义的扫描，提升扫描速度

    新特征:
        1. 主机端口扫描过程中的存活判定，提高多端口的扫描速度
           单个主机的扫描，如果出现过多的 filtered 则会自动放弃该主机的扫描,
           当主机有端口出现 closed/open 状态时，则会强制所有指定的端口扫描
           filtered 过多的阈值设置 `-a` 默认 `1024`
           关闭自动丢弃机制，强制扫描 `-A`
        2. -ap 追加默认端口参数
        3. -c 允许 closed 状态打印输出 (仅 TCP)
    增强:
        1. TCP Connect 中修复没有路由可达的情况下，放弃该主机的扫描
        2. TCP Connect 中修复网络地址 `.0/.255` 等不可达情况下，放弃该主机的扫描
        3. 对位置的错误情况，添加了更详细的错误输出，便于调试
        4. 增强扫描结束后统计的数据显示
        5. 降低了耗时统计的时间单位精确度
        6. 对端口列表进行去重处理
        7. 对端口随机后还是会优先探测常用端口列表 (Targets:Ports 则常用端口交集),
           从而提高自动放弃扫描机制可靠性
    修复:
        1. 使用 Targets:Ports 时，统计的任务数有误问题
        5. 使用 Targets:Ports 时，不会随机化端口扫描顺序的问题


### v1.0.0
    初次发布
