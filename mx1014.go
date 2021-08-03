package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)


type Target struct {
   host string
   ports []string
}


func ErrPrint(msg string) {
    log.Printf("[!] %s\n", msg)
    os.Exit(1)
}

func secondToTime(second int) string {
    minute := second / 60
    if minute == 0 {
        return fmt.Sprintf("%ds", second)
    }else{
        return fmt.Sprintf("%dm%ds", minute, second % 60)
    }
}

func Shuffle(vals []string) []string {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  ret := make([]string, len(vals))
  perm := r.Perm(len(vals))
  for i, randIndex := range perm {
    ret[i] = vals[randIndex]
  }
  return ret
}


func ShuffleTarget(vals []Target) []Target {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  ret := make([]Target, len(vals))
  perm := r.Perm(len(vals))
  for i, randIndex := range perm {
    ret[i] = vals[randIndex]
  }
  return ret
}


func RemoveRepeatedElement(arr []string) []string {
    var newArr []string
    set := make(map[string]bool)
    for _, element := range arr {
        repeat := set[element]
        if !repeat {
            newArr = append(newArr, element)
            set[element] = true
        }
    }
    return newArr
}


func inc(ip net.IP) {
    for i := len(ip) - 1; i >= 0; i-- {
        ip[i]++
        if ip[i] > 0 {
            break
        }
    }
}


func IPCIDR(cidr string) ([]string, error) {
    var hosts []string
    ip, ipnet, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, err
    }
    for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
        hosts = append(hosts, ip.String())
    }
    return hosts, nil
}


func IPWildcard(target string) ([]string, error) {
    var hosts []string
    items := strings.Split(target, ".")
    var blocks [4][]string
    for i := 0; i <= 3; i++ {
        var block []string
        item := items[i]
        if item == "*" {
            for j := 0; j < 256; j++ {
                block = append(block, strconv.Itoa(j))
            }
        } else if strings.ContainsAny(item, "-") {
            a := strings.Split(item, "-")
            Start, err := strconv.Atoi(a[0])
            if err != nil {
                return nil, err
            }
            End, err := strconv.Atoi(a[1])
            if err != nil {
                return nil, err
            }
            if Start >= End {
                return nil, err
            }
            for j := Start; j <= End; j++ {
                block = append(block, strconv.Itoa(j))
            }
        } else {
            j, err := strconv.Atoi(item)
            if err != nil {
                return nil, err
            }
            block = append(block, strconv.Itoa(j))
        }
        blocks[i] = block
    }
    for _, a1 := range(blocks[0]) {
        for _, a2 := range(blocks[1]) {
            for _, a3 := range(blocks[2]) {
                for _, a4 := range(blocks[3]) {
                    items := [4]string{a1, a2, a3, a4}
                    ip := strings.Join(items[:], ".")
                    hosts = append(hosts, ip)
                }
            }
        }
    }
    return hosts, nil
}


func IsIP(str string) (bool){
    return strings.Count(str, ".") == 3 &&
           !strings.ContainsAny(strings.ToUpper(str), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}


func ParsePortRange(portList string) ([]string) {
    var ports []string
    portList2 := strings.Split(portList, ",")

    for _, i := range portList2 {
        if strings.ContainsAny(i, "-") {
            a := strings.Split(i, "-")
            startPort, err := strconv.Atoi(a[0])
            if err != nil {
                startPort = 1
            } else if startPort < 1 {
                ErrPrint("StartPort strconv error")
            }
            endPort, err := strconv.Atoi(a[1])
            if err != nil {
                endPort = 65535
            } else if endPort > 65535 {
                ErrPrint("EndPort strconv error")
            }
            for j := startPort; j <= endPort; j++ {
                ports = append(ports, strconv.Itoa(j))
            }
        } else {
            singlePort, err := strconv.Atoi(i)
            if err != nil {
                ErrPrint("SinglePort strconv error")
            }
            ports = append(ports, strconv.Itoa(singlePort))
        }
    }
    return ports
}


func ParseTarget(target string) ([]Target, error) {
    var targets []Target
    var ports []string
    var portsLen int

    if strings.ContainsAny(target, ":") {
        items := strings.Split(target, ":")
        target = items[0]
        ports = ParsePortRange(items[1])
        if !order {
            ports = Shuffle(ports)
        }
        ports = AdjustPortsList(ports)
        portsLen = len(ports)
    } else {
        portsLen = defaultPortsLen
    }
    if strings.ContainsAny(target, "/") {
        hosts, err := IPCIDR(target)
        if err != nil {
            return nil, err
        }
        for _, host := range(hosts) {
            targets = append(targets, Target{host: host, ports: ports})
        }
    } else if IsIP(target) && strings.ContainsAny(target, "*-") {
        hosts, err := IPWildcard(target)
        if err != nil {
            return nil, err
        }
        for _, host := range(hosts) {
            targets = append(targets, Target{host: host, ports: ports})
        }
    } else {
        hosts := []string{ target }
        for _, host := range(hosts) {
            targets = append(targets, Target{host: host, ports: ports})
        }
    }
    total += portsLen * len(targets)
    return targets, nil
}


// return open: 0, closed: 1, filtered: 2, noroute: 3, denied: 4, down: 5, unkown: -1
func TcpConnect(targetAddr string) int {
    conn, err := net.DialTimeout("tcp", targetAddr, time.Millisecond*time.Duration(timeout))
    if err != nil {
        errMsg := err.Error()
        if strings.Contains(errMsg, "refused") {
            return 1
        } else if strings.Contains(errMsg, "timeout") {
            return 2
        } else if strings.Contains(errMsg, "no route to host") {
            return 3
        } else if strings.Contains(errMsg, "permission denied") {
            return 4
        } else if strings.Contains(errMsg, "host is down") {
            return 5
        } else {
            log.Printf("# [Unkown!!!] %s => %s", targetAddr, err)
            return -1
        }
    }
    defer conn.Close()
    if echoMode {
        port := strings.Split(targetAddr, ":")[1]
        msg := strings.Replace(senddata, "%port%", port, -1)
        conn.Write([]byte(msg))
    }
    return 0
}


func UdpConnect(targetAddr string) bool {
    conn, err := net.DialTimeout("udp", targetAddr, time.Millisecond*time.Duration(timeout))
    if err != nil {
        errMsg := err.Error()
        if verbose {
            log.Printf("# Error: %s (%s)\n", targetAddr, errMsg)
        }
        return false
    }
    defer conn.Close()
    port := strings.Split(targetAddr, ":")[1]
    msg := strings.Replace(senddata, "%port%", port, -1)
    conn.Write([]byte(msg))
    return true
}


func portScan(targets []Target, dports []string) int {
    wg := sync.WaitGroup{}
    targetsChan := make(chan string, timeout)
    poolCount := numOfgoroutine

    go func() {
        for {
            time.Sleep(time.Second * time.Duration(progressDelay))
            rate := float64(doneCount) * 100 / float64(total)
            second := time.Since(startTime).Seconds()
            pps := float64(doneCount) / second
            remaining := second * 100 / float64(rate) - second
            remainingTime := secondToTime(int(remaining))
            log.Printf("# Progress (%d/%d) open: %d, pps: %.0f, rate: %0.f%% (RD %s)\n", doneCount, total, openCount, pps, rate, remainingTime)
        }
    }()


    for i := 0; i <= poolCount; i++ {
        go func() {
            for targetAddr := range targetsChan {
                if udpmode {
                    UdpConnect(targetAddr)
                } else {
                    mutex.Lock()
                    host := strings.Split(targetAddr, ":")[0]
                    var filterCount int
                    if forceScan {
                        filterCount = 65536
                    } else {
                        filterCount = targetFilterCount[host]
                    }
                    mutex.Unlock()
                    // case filterCount
                    // when ...autoDiscard      when continuescan
                    // when autoDiscard...65536 when stopscan
                    // when 65536..             when forcescan
                    if filterCount >= 65536 || filterCount < autoDiscard {
                        flag := TcpConnect(targetAddr)
                        mutex.Lock()
                        switch flag {
                        case 0: //open
                            targetFilterCount[host] = 65536
                            openCount++
                            if aliveMode {
                                log.Print(host)
                            } else {
                                log.Print(targetAddr)
                            }
                        case 1: //closed
                            targetFilterCount[host] = 65536
                            if aliveMode {
                                log.Print(host)
                            } else if verbose || closedMode {
                                fmt.Printf("# closed: %s\n", targetAddr)
                            }
                        case 2: //filtered
                            if filterCount < 65536 {
                                targetFilterCount[host]++
                            }
                            if verbose {
                                fmt.Printf("# filtered: %s\n", targetAddr)
                            }
                        case 3: //noroute
                            targetFilterCount[host] = autoDiscard
                            if verbose {
                                log.Printf("# %s no route to host, discard the host\n", host)
                            }
                        case 4: //denied
                            targetFilterCount[host] = autoDiscard
                        case 5: //down
                            targetFilterCount[host] = autoDiscard
                        case -1: //unkown
                        }
                        mutex.Unlock()
                    }
                }
                mutex.Lock()
                doneCount++
                mutex.Unlock()
                wg.Done()
            }
        }()
    }

    for _, target := range targets {
        host  := target.host
        ports := target.ports
        if len(ports) == 0 {
            ports = dports
        }
        for _, port := range ports {
            tcpTarget := host + ":" + port
            targetsChan <- tcpTarget
            wg.Add(1)
        }
    }

    wg.Wait()
    return 0
}

func GetObjectMap(portsList []string) map[string]bool {
    portsMap := make(map[string]bool)
    for _, i := range portsList {
        portsMap[i] = true
    }
    return portsMap
}

func AdjustPortsList(portsList []string) []string {
    var resList []string
    for _, i := range portsList {
        if commonPortsMap[i] {
            resList = append(resList, i)
        }
    }
    resList = append(resList, portsList...)
    resList = RemoveRepeatedElement(resList)
    return resList
}

var (
    portRanges      string
    addPort         string
    numOfgoroutine  int
    outfile         string
    infile          string
    timeout         int
    autoDiscard     int
    order           bool
    verbose         bool
    udpmode         bool
    forceScan       bool
    echoMode        bool
    closedMode      bool
    showPorts       bool
    aliveMode       bool
    senddata        string
    total           int
    openCount       int
    doneCount       int
    defaultPortsLen int
    progressDelay   int
    mutex           sync.Mutex
    startTime       time.Time

    targetFilterCount = make(map[string]int)
    rawCommonPorts    = "22,80,81,82,83,84,85,86,88,89,90,99,135,137,138,139,389,443,445,800,801,808,880,888,889,1000,1010,1024,1080,1433,1521,1980,3000,3128,3308,3389,3505,4430,4433,4560,5432,5555,5800,5900,5985,5986,6080,6379,6588,6677,6868,7000,7001,7002,7003,7005,7007,7070,7080,7200,7777,7890,8000,8001,8002,8003,8004,8006,8008,8010,8011,8012,8016,8020,8053,8060,8070,8080,8081,8082,8083,8084,8085,8086,8087,8088,8089,8090,8091,8099,8100,8161,8180,8181,8182,8200,8280,8300,8360,8443,8484,8800,8880,8881,8888,8899,8989,9000,9001,9002,9043,9060,9080,9081,9085,9090,9091,9200,9875,9999,10000,18080,28017,38501,38888,41516"
    commonPorts       = ParsePortRange(rawCommonPorts)
    commonPortsMap    = GetObjectMap(commonPorts)
)


func usage() {
    fmt.Fprintf(
        os.Stderr, `
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
`)
    flagSet := flag.CommandLine
    optsOrder := []string{"p", "ap", "i", "t", "T", "o", "r", "u", "e", "c", "d", "D", "l", "a", "A", "v", "sp"}
    for _, name := range optsOrder {
        fl4g := flagSet.Lookup(name)
        fmt.Printf("    -%s", fl4g.Name)
        fmt.Printf(" %s\n", fl4g.Usage)
    }
    os.Exit(0)
}


func init() {
    flag.StringVar(&portRanges,  "p", rawCommonPorts, " Ports  Default port ranges. (Default is common ports)")
    flag.StringVar(&addPort,     "ap", "",            "Ports  Append default ports")
    flag.IntVar(&numOfgoroutine, "t", 256,            " Int    The Number of Goroutine (Default is 256)")
    flag.IntVar(&timeout,        "T", 1514,           " Int    TCP Connect Timeout (Default is 1514ms)")
    flag.StringVar(&infile,      "i", "",             " File   Target input from list")
    flag.StringVar(&outfile,     "o", "",             " File   Output file path")
    flag.BoolVar(&order,         "r", false,          "        Scan in import order")
    flag.BoolVar(&udpmode,       "u", false,          "        UDP spray")
    flag.BoolVar(&echoMode,      "e", false,          "        Echo mode (TCP needs to be manually)")
    flag.BoolVar(&closedMode,    "c", false,          "        Allow display of closed ports (Only TCP)")
    flag.IntVar(&autoDiscard,    "a", 1014,           " Int    Too many filtered, Discard the host (Default is 1014)")
    flag.BoolVar(&forceScan,     "A", false,          "        Disable auto disable")
    flag.BoolVar(&aliveMode,     "l", false,          "        Output alive host")
    flag.StringVar(&senddata,    "d", "%port%\n",     " Str    Specify Echo mode data (Default is \"%port%\\n\")")
    flag.IntVar(&progressDelay,  "D", 5,              " Int    Progress Bar Refresh Delay (Default is 5s)")
    flag.BoolVar(&verbose,       "v", false,          "        Verbose mode")
    flag.BoolVar(&showPorts,     "sp", false,         "       Only show default ports")
    flag.Usage = usage
}


func main() {
    total = 0
    openCount = 0
    startTime = time.Now()
    flag.Parse()

    if showPorts {
        fmt.Printf("Count: %d\n", len(commonPorts))
        fmt.Println(rawCommonPorts)
        os.Exit(0)
    }

    if addPort != "" {
        portRanges += ( "," + addPort )
    }
    defaultPorts := ParsePortRange(portRanges)
    if !order {
        defaultPorts = Shuffle(defaultPorts)
    }

    defaultPorts = AdjustPortsList(defaultPorts)
    defaultPortsLen = len(defaultPorts)

    var rawTargets []string
    var allTargets []Target
    if infile == "" {
        rawTargets = flag.Args()
    } else {
        file, err := os.Open(infile)
        if err != nil {
            ErrPrint(fmt.Sprintf("File read failed: %s", infile))
        }
        defer file.Close()
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            line := strings.Trim(scanner.Text(), " \t\f\v")
            if line != "" {
                rawTargets = append(rawTargets, line)
            }
        }
    }

    for _, rawTarget := range rawTargets {
        targets, err := ParseTarget(rawTarget)
        if err != nil {
            ErrPrint(fmt.Sprintf("Wrong target: %s", rawTarget))
        }
        for _, target := range targets {
            allTargets = append(allTargets, target)
        }
    }

    if !order {
        allTargets   = ShuffleTarget(allTargets)
    }

    if len(allTargets) == 0 {
        flag.Usage()
    }

    log.SetFlags(0)
    if outfile != "" {
        logFile, err := os.OpenFile(outfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, os.ModeAppend | os.ModePerm)
        if err != nil {
            ErrPrint("Open output file failed")
        }

        defer logFile.Close()
        out := io.MultiWriter(os.Stdout, logFile)
        log.SetOutput(out)
    }

    allTargetsSize := len(allTargets)
    EchoModePrompt := ""
    if echoMode && !udpmode {
        EchoModePrompt = " (TCP Echo)"
    }
    log.Printf("# %s Start scanning %d hosts...%s\n\n", startTime.Format("2006/01/02 15:04:05"), allTargetsSize, EchoModePrompt)
    portScan(allTargets, defaultPorts)
    spendTime := time.Since(startTime).Seconds()
    pps := float64(total) / spendTime
    hostAlive := 0
    for _, i := range targetFilterCount {
        if i >= 65536 {
            hostAlive++
        }
    }
    aliveRate := hostAlive * 100.0 / allTargetsSize
    endTime := time.Now().Format("2006/01/02 15:04:05")
    log.Printf("\n# %s Finished %d tasks. alive: %d%% (%d/%d), open: %d, pps: %.0f, time: %s\n", endTime, total, aliveRate, hostAlive, allTargetsSize, openCount, pps, secondToTime(int(spendTime)))
}
