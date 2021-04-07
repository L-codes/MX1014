package main


import (
    "fmt"
    "flag"
    "io"
    "log"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "math/rand"
    "time"
    "bufio"
)


type Target struct {
   host string
   ports []string
}


func ErrPrint(msg string) {
    log.Printf("[!] %s\n", msg)
    os.Exit(1)
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


func inc(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}


func CIDR(cidr string) ([]string, error) {
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
                ErrPrint("StartPort strconv error")
            }
            endPort, err := strconv.Atoi(a[1])
            if err != nil {
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
    if strings.ContainsAny(target, ":") {
        items := strings.Split(target, ":")
        target = items[0]
        ports = ParsePortRange(items[1])
        total += len(ports)
    } else {
        total += defaultPortsLen
    }
    if strings.ContainsAny(target, "/") {
        hosts, err := CIDR(target)
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
    return targets, nil
}


func TcpConnect(targetAddr string) bool {
    conn, err := net.DialTimeout("tcp", targetAddr, time.Millisecond*time.Duration(timeout))
    if err != nil {
        errMsg := err.Error()
        if verbose {
            if strings.Contains(errMsg, "connection refused") {
                fmt.Printf("# %s closed\n", targetAddr)
            } else if strings.Contains(errMsg, "timeout") {
                fmt.Printf("# %s filtered\n", targetAddr)
            }
        }
        return false
    }
    defer conn.Close()
    if echo {
        port := strings.Split(targetAddr, ":")[1]
        msg := strings.Replace(senddata, "%port%", port, -1)
        conn.Write([]byte(msg))
    }
    return true
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
            rate := doneCount * 100.0 / total
            second := time.Since(startTime).Seconds()
            pps := float64(doneCount) / second
            log.Printf("# Progress (%d/%d) open: %d, pps: %.0f, rate: %d%%\n", doneCount, total, openCount, pps, rate)
        }
    }()

    for i := 0; i <= poolCount; i++ {
        go func() {
            for targetAddr := range targetsChan {
                if udpmode {
                    UdpConnect(targetAddr)
                } else {
                    openFlag := TcpConnect(targetAddr)
                    if openFlag {
                        mutex.Lock()
                        openCount++
                        log.Print(targetAddr)
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


var (
    portRanges      string
    numOfgoroutine  int
    outfile         string
    infile          string
    timeout         int
    order           bool
    verbose         bool
    udpmode         bool
    echo            bool
    senddata        string
    total           int
    openCount       int
    doneCount       int
    defaultPortsLen int
    progressDelay   int
    mutex           sync.Mutex
    startTime       time.Time
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
`)
    flagSet := flag.CommandLine
    optsOrder := []string{"p", "i", "t", "T", "o", "r", "u", "e", "d", "D", "v"}
    for _, name := range optsOrder {
        fl4g := flagSet.Lookup(name)
        fmt.Printf("    -%s", fl4g.Name)
        fmt.Printf(" %s\n", fl4g.Usage)
    }
    os.Exit(0)
}


func init() {
    default_ports := "22,80,81,82,88,89,135,137,138,139,389,443,445,1080,1433,1521,3128,3308,3389,4430,4433,4560,5432,5800,5900,5985,5986,6379,6588,7001,7002,8000,8001,8002,8009,8161,8080,8081,8082,8090,9000,9090,9043,9060,9200,9875,8443,8880,8888"
    flag.StringVar(&portRanges,  "p", default_ports, "Ports  Default port ranges. (Default is common ports")
    flag.IntVar(&numOfgoroutine, "t", 256,           "Int    The Number of Goroutine (Default is 256)")
    flag.IntVar(&timeout,        "T", 1014,          "Int    TCP Connect Timeout (Default is 1014ms)")
    flag.StringVar(&infile,      "i", "",            "File   Target input from list")
    flag.StringVar(&outfile,     "o", "",            "File   Output file path")
    flag.BoolVar(&order,         "r", false,         "       Scan in import order")
    flag.BoolVar(&udpmode,       "u", false,         "       UDP spray")
    flag.BoolVar(&echo,          "e", false,         "       Echo mode (TCP needs to be manually)")
    flag.StringVar(&senddata,    "d", "%port%\n",    "Str    Specify Echo mode data (Default is \"%port%\\n\")")
    flag.IntVar(&progressDelay,  "D", 5,             "Int    Progress Bar Refresh Delay (Default is 5s)")
    flag.BoolVar(&verbose,       "v", false,         "       Verbose mode")
    flag.Usage = usage
}


func main() {
    total = 0
    openCount = 0
    startTime = time.Now()
    flag.Parse()

    defaultPorts := ParsePortRange(portRanges)
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
        defaultPorts = Shuffle(defaultPorts)
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
    log.Printf("# %s Start scanning %d hosts...\n\n", startTime.Format("2006/01/02 15:04:05"), allTargetsSize)
    portScan(allTargets, defaultPorts)
    spendTime := time.Since(startTime).Seconds()
    pps := float64(total) / spendTime
    log.Printf("\n# Finished. host: %d, task: %d, open: %d, pps: %.0f, time: %.2fs\n", allTargetsSize, total, openCount, pps, spendTime)
}
