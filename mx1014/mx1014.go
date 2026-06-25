package mx1014

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

func ErrPrint(msg string) {
    log.Printf("[!] %s\n", msg)
    os.Exit(1)
}

func secondToTime(second uint64) string {
    day := second / 86400
    hour := (second % 86400) / 3600
    minute := (second % 3600) / 60
    if day != 0 {
        return fmt.Sprintf("%dd%dh%dm%ds", day, hour, minute, second%60)
    } else if hour != 0 {
        return fmt.Sprintf("%dh%dm%ds", hour, minute, second%60)
    } else if minute != 0 {
        return fmt.Sprintf("%dm%ds", minute, second%60)
    } else {
        return fmt.Sprintf("%ds", second)
    }
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
    size := len(hosts)

    if size > 2 {
        hosts = hosts[1 : size-1]
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
    for _, a1 := range blocks[0] {
        for _, a2 := range blocks[1] {
            for _, a3 := range blocks[2] {
                for _, a4 := range blocks[3] {
                    items := [4]string{a1, a2, a3, a4}
                    ip := strings.Join(items[:], ".")
                    hosts = append(hosts, ip)
                }
            }
        }
    }
    return hosts, nil
}

func IsIP(str string) bool {
    return strings.Count(str, ".") == 3 &&
        !strings.ContainsAny(strings.ToUpper(str), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func AddFuzzPort(ports []string) []string {
    var fuzzPorts []string
    for _, port := range ports {
        portNum, _ := strconv.Atoi(port)
        if len(port) == 2 {
            fuzzPorts = append(fuzzPorts, port+port)
        } else if len(port) == 4 {
            // left
            for i := 1; i <= 6; i++ {
                fuzzPortNum := portNum + i*10000
                if fuzzPortNum > 65535 {
                    break
                }
                fuzzPorts = append(fuzzPorts, strconv.Itoa(fuzzPortNum))
            }
            // right
            for i := 0; i <= 9; i++ {
                fuzzPortNum := portNum*10 + i
                if fuzzPortNum > 65535 {
                    break
                }
                fuzzPorts = append(fuzzPorts, strconv.Itoa(fuzzPortNum))
            }
        }
        // side
        if portNum < 65535 {
            fuzzPorts = append(fuzzPorts, strconv.Itoa(portNum+1))
        }
        if portNum > 1 {
            fuzzPorts = append(fuzzPorts, strconv.Itoa(portNum-1))
        }
        if len(port) <= 4 {
            // left overlapping
            leftOverlapping := string(port[0]) + port
            leftOverlappingNum, _ := strconv.Atoi(leftOverlapping)
            if leftOverlappingNum <= 65535 {
                fuzzPorts = append(fuzzPorts, leftOverlapping)
            }
            // right overlapping
            rightOverlapping := port + string(port[len(port)-1])
            rightOverlappingNum, _ := strconv.Atoi(rightOverlapping)
            if rightOverlappingNum <= 65535 {
                fuzzPorts = append(fuzzPorts, rightOverlapping)
            }
        }
    }
    fuzzPorts = append(fuzzPorts, ports...)
    return fuzzPorts
}

func ParsePortRange(portList string, ignoreFuzz bool) []string {
    var ports []string
    portList2 := strings.Split(portList, ",")

    for _, i := range portList2 {
        if portAlias := portGroup[i]; portAlias != nil {
            for _, port := range portAlias {
                ports = append(ports, strconv.Itoa(port))
            }
        } else if strings.ContainsAny(i, "-") {
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
            if singlePort > 65535 || singlePort <= 0 {
                ErrPrint("Wrong port number: " + i)
            }
            ports = append(ports, strconv.Itoa(singlePort))
        }
    }
    if !ignoreFuzz && fuzzPort {
        ports = AddFuzzPort(ports)
    }
    ports = RemoveRepeatedElement(ports)
    return ports
}

func ParseTarget(target string, defaultPorts []string) error {
    var ports []string
    var portsLen uint64

    if strings.ContainsAny(target, ":") {
        items := strings.Split(target, ":")
        target = items[0]
        ports = ParsePortRange(items[1], false)
        portsLen = uint64(len(ports))
    } else {
        ports = defaultPorts
        portsLen = defaultPortsLen
    }

    if strings.ContainsAny(target, "/") {
        hosts, err := IPCIDR(target)
        if err != nil {
            return err
        }
        mutex.Lock()
        hostMap[target] = hosts
        mutex.Unlock()
    } else if IsIP(target) && strings.ContainsAny(target, "*-") {
        hosts, err := IPWildcard(target)
        if err != nil {
            return err
        }
        mutex.Lock()
        hostMap[target] = hosts
        mutex.Unlock()
    } else {
        _, err := net.LookupHost(target)
        if err != nil {
            if target[0] == 0x2d { // "-"
                log.Println("[*] Usage: ./mx1014 [Options] [Target1] [Target2]...")
            }
            return err
        }
        mutex.Lock()
        hostMap[target] = []string{target}
        mutex.Unlock()
    }

    mutex.Lock()
    for _, port := range ports {
        portMap[port] = append(portMap[port], target)
    }

    hostCount := uint64(len(hostMap[target]))
    hostTotal += hostCount
    total += portsLen * hostCount
    mutex.Unlock()

    return nil
}

type socks5Auth struct {
    username string
    password string
}

func socks5Dial(proxyAddr, targetAddr string, auth *socks5Auth, timeout time.Duration) (net.Conn, error) {
    conn, err := net.DialTimeout("tcp", proxyAddr, timeout)
    if err != nil {
        return nil, err
    }

    deadline := time.Now().Add(timeout)
    conn.SetDeadline(deadline)

    // Negotiate auth method
    var msg []byte
    if auth != nil {
        msg = []byte{0x05, 0x02, 0x00, 0x02}
    } else {
        msg = []byte{0x05, 0x01, 0x00}
    }
    if _, err := conn.Write(msg); err != nil {
        conn.Close()
        return nil, err
    }

    resp := make([]byte, 2)
    if _, err := io.ReadFull(conn, resp); err != nil {
        conn.Close()
        return nil, err
    }
    if resp[0] != 0x05 {
        conn.Close()
        return nil, fmt.Errorf("socks5: unsupported version %d", resp[0])
    }

    if resp[1] == 0x02 {
        if auth == nil {
            conn.Close()
            return nil, fmt.Errorf("socks5: server requires authentication")
        }
        // RFC 1929 username/password auth
        authMsg := []byte{0x01, byte(len(auth.username))}
        authMsg = append(authMsg, []byte(auth.username)...)
        authMsg = append(authMsg, byte(len(auth.password)))
        authMsg = append(authMsg, []byte(auth.password)...)
        if _, err := conn.Write(authMsg); err != nil {
            conn.Close()
            return nil, err
        }
        authResp := make([]byte, 2)
        if _, err := io.ReadFull(conn, authResp); err != nil {
            conn.Close()
            return nil, err
        }
        if authResp[1] != 0x00 {
            conn.Close()
            return nil, fmt.Errorf("socks5: authentication failed")
        }
    } else if resp[1] != 0x00 {
        conn.Close()
        return nil, fmt.Errorf("socks5: unsupported auth method %d", resp[1])
    }

    // Build connect request
    host, portStr, err := net.SplitHostPort(targetAddr)
    if err != nil {
        conn.Close()
        return nil, err
    }
    port, err := strconv.Atoi(portStr)
    if err != nil {
        conn.Close()
        return nil, err
    }

    var req []byte
    req = append(req, 0x05, 0x01, 0x00) // ver, cmd(connect), rsv

    if ip := net.ParseIP(host); ip != nil {
        if ip4 := ip.To4(); ip4 != nil {
            req = append(req, 0x01)
            req = append(req, ip4...)
        } else {
            req = append(req, 0x04)
            req = append(req, ip.To16()...)
        }
    } else {
        req = append(req, 0x03, byte(len(host)))
        req = append(req, []byte(host)...)
    }
    req = append(req, byte(port>>8), byte(port))

    if _, err := conn.Write(req); err != nil {
        conn.Close()
        return nil, err
    }

    // Read response header
    respHeader := make([]byte, 4)
    if _, err := io.ReadFull(conn, respHeader); err != nil {
        conn.Close()
        return nil, err
    }
    if respHeader[0] != 0x05 {
        conn.Close()
        return nil, fmt.Errorf("socks5: response version %d", respHeader[0])
    }

    // Translate SOCKS5 response codes
    switch respHeader[1] {
    case 0x00: // success
    case 0x05: // connection refused
        conn.Close()
        return nil, fmt.Errorf("connection refused")
    case 0x03: // network unreachable
        conn.Close()
        return nil, fmt.Errorf("network is unreachable")
    case 0x04: // host unreachable
        conn.Close()
        return nil, fmt.Errorf("no route to host")
    default:
        conn.Close()
        return nil, fmt.Errorf("socks5: request failed (code %d)", respHeader[1])
    }

    // Read remaining address/port (variable length)
    atyp := respHeader[3]
    var addrLen int
    switch atyp {
    case 0x01:
        addrLen = 4
    case 0x03:
        lenByte := make([]byte, 1)
        if _, err := io.ReadFull(conn, lenByte); err != nil {
            conn.Close()
            return nil, err
        }
        addrLen = int(lenByte[0])
    case 0x04:
        addrLen = 16
    default:
        conn.Close()
        return nil, fmt.Errorf("socks5: unsupported address type %d", atyp)
    }
    addrBytes := make([]byte, addrLen+2) // addr + port
    if _, err := io.ReadFull(conn, addrBytes); err != nil {
        conn.Close()
        return nil, err
    }

    conn.SetDeadline(time.Time{})
    return conn, nil
}

// return open: 0, closed: 1, filtered: 2, noroute: 3, denied: 4, down: 5, error_host: 6, unkown: -1, abort: -2
func TcpConnect(targetAddr string) int {
    var conn net.Conn
    var err error

    if proxy != "" {
        conn, err = socks5Dial(proxyAddr, targetAddr, proxyAuth, time.Millisecond*time.Duration(timeout))
    } else {
        conn, err = net.DialTimeout("tcp", targetAddr, time.Millisecond*time.Duration(timeout))
    }
    if err != nil {
        errMsg := err.Error()
        if strings.Contains(errMsg, "refused") {
            return 1
        } else if strings.Contains(errMsg, "An attempt was made to access a socket in a way forbidden by its access permissions.") {
            return 1
        } else if strings.Contains(errMsg, "timeout") {
            return 2
        } else if strings.Contains(errMsg, "protocol not available") {
            return 2
        } else if strings.Contains(errMsg, "no route to host") {
            return 3
        } else if strings.Contains(errMsg, "permission denied") {
            return 4
        } else if strings.Contains(errMsg, "host is down") {
            return 5
        } else if strings.Contains(errMsg, "no such host") {
            return 6
        } else if strings.Contains(errMsg, "network is unreachable") {
            return 6
        } else if strings.Contains(errMsg, "The requested address is not valid in its context.") {
            return 6
        } else if strings.Contains(errMsg, "A socket operation was attempted to an unreachable") {
            return 6
        } else if strings.Contains(errMsg, "too many open files") {
            return -2
        } else if proxy != "" {
			if strings.Contains(errMsg, "EOF") {
				return 1
			} else if strings.Contains(errMsg, "socks5: authentication failed") {
				return -2
			} else {
				log.Printf("# [Unkown!!!] %s => %s", targetAddr, err)
				return -1
			}
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

func UdpConnect(targetAddr string) int {
    conn, err := net.DialTimeout("udp", targetAddr, time.Millisecond*time.Duration(timeout))
    if err != nil {
        errMsg := err.Error()
        if verbose {
            log.Printf("# Error: %s (%s)\n", targetAddr, errMsg)
        }
        return 0
    }
    defer conn.Close()
    port := strings.Split(targetAddr, ":")[1]
    msg := strings.Replace(senddata, "%port%", port, -1)
    conn.Write([]byte(msg))
    return 1
}

func ProgressBar() {
    doneCount = 0
    for {
        time.Sleep(time.Second * time.Duration(progressDelay))
        rate := float64(doneCount) * 100 / float64(total)
        second := time.Since(startTime).Seconds()
        pps := float64(doneCount) / second
        remaining := second*100/float64(rate) - second
        remainingTime := secondToTime(uint64(remaining))
        log.Printf("# Progress (%d/%d) up: %d, open: %d, discard: %d, pps: %.0f, rate: %0.f%% (RD %s)\n", doneCount, total, hostUpCount, openCount, hostDiscard, pps, rate, remainingTime)
    }
}

func SendPacket(targetAddr string) {
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
                if targetFilterCount[host] < 65536 { // First found alive
                    hostUpCount++
                    targetFilterCount[host] = 65536
                }
                openCount++
                if aliveMode {
                    log.Print(host)
                } else {
                    port := strings.Split(targetAddr, ":")[1]
                    servers := portServersMap[port]
                    if disableProtocolName || servers == "" {
                        log.Print(targetAddr)
                    } else {
                        log.Printf("%-26s (%s)", targetAddr, servers)
                    }
                }
            case 1: //closed
                if targetFilterCount[host] < 65536 { // First found alive
                    hostUpCount++
                    targetFilterCount[host] = 65536
                }
                if aliveMode {
                    log.Print(host)
                } else if verbose || closedMode {
                    fmt.Printf("# closed: %s\n", targetAddr)
                }
            case 2: //filtered
                if filterCount < 65536 {
                    targetFilterCount[host]++
                    if targetFilterCount[host] == autoDiscard { // Just met max
                        hostDiscard++
                    }
                }
                if verbose {
                    fmt.Printf("# filtered: %s\n", targetAddr)
                }
            case 3: //noroute
                targetFilterCount[host] = autoDiscard + 1
                if verbose {
                    log.Printf("# %s no route to host, discard the host\n", host)
                }
            case 4: //denied
                targetFilterCount[host] = autoDiscard + 1
            case 5: //down
                targetFilterCount[host] = autoDiscard + 1
            case 6: //error_host
                targetFilterCount[host] = autoDiscard + 1
            case -2: //abort
                log.Printf("# too many open files !!!")
                log.Printf("# Please lower the `-t` value and run again")
                os.Exit(-2)
            case -1: //unkown
            }
            mutex.Unlock()
        }
    }
}

func RejectAllOpenProgressBar() {
    doneCount = 0
    testTotal := hostTotal * uint64(rejectAllOpenTimes)
    stopRejectAllOpenProgressBar = false
    for {
        time.Sleep(time.Second * time.Duration(progressDelay))
        if stopRejectAllOpenProgressBar == true {
            break
        }
        rate := float64(doneCount) * 100 / float64(testTotal)
        second := time.Since(startTime).Seconds()
        pps := float64(doneCount) / second
        remaining := second*100/float64(rate) - second
        remainingTime := secondToTime(uint64(remaining))
        log.Printf("# reject all open (%d/%d) pps: %.0f, rate: %0.f%% (RD %s)\n", doneCount, testTotal, pps, rate, remainingTime)
    }
}

func SendRandTCPPacket(host string) {

    targetAddr := host + ":" + RandPort(50000, 65535)
    flag := TcpConnect(targetAddr)

    mutex.Lock()
    if flag == 0 {
        rejectOpenCount[host]++
    }
    mutex.Unlock()
}

func RandPort(min int, max int) string {
    portNum := rand.Intn(max-min) + min
    return strconv.Itoa(portNum)
}

func RejectAllOpenTargets() {
    wg := sync.WaitGroup{}
    targetsChan := make(chan string, timeout)

    go RejectAllOpenProgressBar()

    for i := 0; i <= numOfgoroutine; i++ {
        go func() {
            for host := range targetsChan {
                SendRandTCPPacket(host)
                mutex.Lock()
                doneCount++
                mutex.Unlock()
                wg.Done()
            }
        }()
    }

    for _, hosts := range hostMap {
        for _, host := range hosts {
            for j := 0; j < rejectAllOpenTimes; j++ {
                targetsChan <- host
                wg.Add(1)
            }
        }
    }
    wg.Wait()

    stopRejectAllOpenProgressBar = true

    for host, openCount := range rejectOpenCount {
        if openCount == rejectAllOpenTimes {
            rejectCount++
            log.Printf("# reject all open target: %s\n", host)
        }
    }
}

func PortScan() {
    wg := sync.WaitGroup{}
    targetsChan := make(chan string, timeout)

    go ProgressBar()

    for i := 0; i <= numOfgoroutine; i++ {
        go func() {
            for targetAddr := range targetsChan {
                SendPacket(targetAddr)
                mutex.Lock()
                doneCount++
                mutex.Unlock()
                wg.Done()
            }
        }()
    }

    if headPortRanges != "" {
        for _, port := range ParsePortRange(headPortRanges, true) {
            rawTargets := portMap[port]
            for _, rawTarget := range rawTargets {
                for _, host := range hostMap[rawTarget] {
                    if rejectOpenCount[host] != rejectAllOpenTimes {
                        targetAddr := host + ":" + port
                        targetsChan <- targetAddr
                        wg.Add(1)
                    }
                }
            }
            delete(portMap, port)
        }
    }

    for port, rawTargets := range portMap {
        for _, rawTarget := range rawTargets {
            for _, host := range hostMap[rawTarget] {
                if rejectOpenCount[host] != rejectAllOpenTimes {
                    targetAddr := host + ":" + port
                    targetsChan <- targetAddr
                    wg.Add(1)
                }
            }
        }
    }
    wg.Wait()
}

func GetObjectMap(portsList []string) map[string]bool {
    portsMap := make(map[string]bool)
    for _, i := range portsList {
        portsMap[i] = true
    }
    return portsMap
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

func FileReadlines(readfile string) []string {
    var lines []string
    file, err := os.Open(readfile)
    if err != nil {
        ErrPrint(fmt.Sprintf("File read failed: %s", readfile))
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.Trim(scanner.Text(), " \t\f\v")
        if line != "" && line[0] != 0x23 { // 0x23 == #
            lines = append(lines, line)
        }
    }
    return lines
}

var (
    // args
    portRanges          string
    numOfgoroutine      int
    outfile             string
    infile              string
    timeout             int
    autoDiscard         int
    verbose             bool
    udpmode             bool
    forceScan           bool
    echoMode            bool
    closedMode          bool
    showPorts           bool
    showHosts           bool
    aliveMode           bool
    fuzzPort            bool
    cNet                bool
    ignoreErrHost       bool
    senddata            string
    doneCount           int
    progressDelay       int
    excludePortRanges   string
    excludePorts        []int
    headPortRanges      string
    gatewayRanges       string
    disableProtocolName bool

    proxy               string
    proxyAddr           string
    proxyAuth           *socks5Auth

    stopRejectAllOpenProgressBar bool
    rejectAllOpen                bool
    rejectAllOpenTimes           int
    rejectCount                  int
    rejectOpenCount              = make(map[string]int)

    defaultPortsLen uint64
    mutex           sync.Mutex

    total uint64      = 0
    hostUpCount       = 0
    hostDiscard       = 0
    hostTotal uint64   = 0
    openCount         = 0
    startTime         = time.Now()
    portMap           = make(map[string][]string) // port: rawtargets
    hostMap           = make(map[string][]string) // rawtarget: hosts
    targetFilterCount = make(map[string]int)
    portGroup = map[string][]int {
      "in": []int{ 80,443,8080,22,2222,139,445,5985,5986,873,5800,5900,5901,6379,63790,3389,1433,1434,210,1158,1521,3306,3307,3308,5432,11211,27017,28017,7687,23,25,465,587,2525,109,110,995,143,993,88,389,636,3268,3269,1080,21,115,2121,554,8554,3999,5000,5005,8000,8453,8787,8788,9001,12001,12002,18000,45000,45001,8009,2375,2376,2377,8081,10252,6443,10256,31442,4149,10248,10250,10255,6781,6782,6783,2379,2380,9875,5480,7848,8848,9848,9849,8075,9876,10909,10911,10912,512,513,514,1098,4444,4445,8083,1028,1090,11099,47001,10999,1099,1000,1001,1100,1101,5001,9999,10001,19001,2049,2100,3128,4786,4848,4990,5555,5556,5858,9229,8069,10050,8161,61616,8383,8500,8983,9000,9092,9200,9300,515,3632,4369,623,4712,20880,20881,4446,4447,4457,1111,8443,45566,48620,81,82,83,84,85,86,87,89,90,444,800,801,1024,1443,2000,2001,3001,4430,4433,4443,6000,6001,6002,6003,6080,6588,6666,6888,7004,7005,7006,7007,7008,7009,7080,7443,7777,8001,8002,8003,8004,8005,8006,8007,8008,8010,8011,8012,8013,8014,8015,8016,8017,8018,8019,8020,8021,8022,8023,8024,8025,8026,8027,8028,8029,8030,8040,8050,8060,8066,8070,8082,8084,8085,8086,8087,8088,8089,8090,8091,8092,8093,8094,8095,8096,8097,8098,8099,8100,8101,8102,8103,8104,8105,8106,8107,8108,8109,8110,8111,8181,8182,8200,8282,8363,8761,8800,8866,8873,8881,8882,8883,8884,8885,8886,8887,8888,8889,8890,8899,8900,8989,8999,9002,9003,9004,9005,9006,9007,9008,9009,9010,9099,10000,10080,10800,18080,18090,5003,888,6868,55555,10443,7088,880,12580,12590,5601,7788,3339,15672,8180,7000,7001,7002,7003,7010,7070,7071,8880,9043,9080,9081,9082,9083,9090,9443,3000,9990,161,1352,2171,2181,2888,3888,5632,8041,8042,8480,8485,10003,14000,19888,41414,50010,50020,50030,50060,50070,50075,50090,50470,50475,60010,60030,264,3260,3299,3690,111,1026,5672,135,593,46888,4100,5984,16000,16010,16020,16030,9042,9160,54321,5236,49152,20000,502,102,44818,1962,5007,9600,789,1200,2404,20547 },
      "rce": []int{ 6379,63790,3999,5000,5005,8000,8080,8453,8787,8788,9001,12001,12002,18000,45000,45001,8009,2375,2376,2377,8081,10252,6443,10256,31442,4149,10248,10250,10255,6781,6782,6783,2379,2380,9875,5480,7848,8848,9848,9849,8075,9876,10909,10911,10912,512,513,514,1098,4444,4445,8083,1028,1090,11099,47001,10999,1099,1000,1001,1100,1101,5001,9999,10001,19001,2049,2100,3128,4786,4848,4990,5555,5556,5800,5900,5901,5858,9229,8069,10050,8161,61616,8383,8500,8983,9000,9092,9200,9300,515,3632,4369,623,4712,20880,20881,4446,4447,4457,80,1111,8443,45566,48620 },
      "info": []int{ 21,115,2121,25,465,587,2525,109,110,995,143,993,161,873,1352,2171,2181,2888,3888,5601,5632,8020,8040,8041,8042,8088,8480,8485,9000,9083,10000,10003,14000,19888,41414,50010,50020,50030,50060,50070,50075,50090,50470,50475,60010,60030,264,3260,3299,3690,111,1026,554,8554,5672,135,593,139,3000,46888,1433,1434,210,1158,1521,3306,3307,3308,5432,6379,63790,11211,27017,28017,7687,4100,5000,5984,16000,16010,16020,16030,9001,9042,9160,54321,5236,49152 },
      "brute": []int{ 22,2222,139,445,5985,5986,873,5800,5900,5901,6379,63790,3389,1433,1434,210,1158,1521,3306,3307,3308,5432,11211,27017,28017,7687,23,25,465,587,2525,109,110,995,143,993,88,389,636,3268,3269,1080,21,115,2121,554,8554 },
      "web1": []int{ 80,443,8080 },
      "web2": []int{ 81,82,83,84,85,86,87,88,89,90,444,800,801,1024,1443,2000,2001,3001,4430,4433,4443,5000,5001,5555,5800,6000,6001,6002,6003,6080,6443,6588,6666,6888,7004,7005,7006,7007,7008,7009,7080,7443,7777,8000,8001,8002,8003,8004,8005,8006,8007,8008,8009,8010,8011,8012,8013,8014,8015,8016,8017,8018,8019,8020,8021,8022,8023,8024,8025,8026,8027,8028,8029,8030,8040,8050,8060,8066,8070,8080,8081,8082,8083,8084,8085,8086,8087,8088,8089,8090,8091,8092,8093,8094,8095,8096,8097,8098,8099,8100,8101,8102,8103,8104,8105,8106,8107,8108,8109,8110,8111,8181,8182,8200,8282,8363,8761,8787,8800,8848,8866,8873,8881,8882,8883,8884,8885,8886,8887,8888,8889,8890,8899,8900,8989,8999,9000,9001,9002,9003,9004,9005,9006,9007,9008,9009,9010,9099,9999,10000,10001,10080,10800,18080,18090,8161,61616,5003,888,6868,55555,10443,7088,880,80,443,47001,4446,4447,4457,1098,4444,4445,1111,8443,45566,12580,12590,5601,7788,3339,15672,8180,1080,8983,3128,7000,7001,7002,7003,7010,7070,7071,8880,9043,9080,9081,9082,9083,9090,9443,3000,9200,9300,8069,10050,9990,7848,9848,9849,8075 },
      "iis": []int{ 80,443,47001 },
      "jboss": []int{ 4446,4447,4457,1098,4444,4445,8083,80,1111,8080,8443,45566 },
      "jboss_rmi": []int{ 1098,4444,4445,8083 },
      "jboss_remoting": []int{ 4446,4447,4457 },
      "zookeeper": []int{ 2171,2181,2888,3888 },
      "dubbo": []int{ 20880,20881 },
      "solr": []int{ 8983 },
      "finereport": []int{ 8075 },
      "websphere_web": []int{ 8880,9043,9080,9081,9082,9083,9090,9443 },
      "websphere": []int{ 8880,9043,9080,9081,9082,9083,9090,9443,2809,5558,5578,7276,7286,9060,9100,9353,9401,9402 },
      "activemq": []int{ 8161,61616 },
      "weblogic": []int{ 7000,7001,7002,7003,7010,7070,7071 },
      "squid": []int{ 3128 },
      "rabbitmq": []int{ 15672 },
      "flink": []int{ 8081 },
      "oracle_web": []int{ 3339 },
      "wildfly": []int{ 9990 },
      "baota": []int{ 888,8888 },
      "fastcgi": []int{ 9000 },
      "kc_aom": []int{ 12580,12590 },
      "kibana": []int{ 5601 },
      "portainer": []int{ 9000 },
      "natshell": []int{ 7788 },
      "elasticsearch": []int{ 9200,9300 },
      "rizhiyi": []int{ 8180 },
      "arl": []int{ 5003 },
      "cassini": []int{ 6868 },
      "dlink": []int{ 55555 },
      "fortigate": []int{ 10443 },
      "nexus": []int{ 8081 },
      "sapido": []int{ 1080 },
      "yapi": []int{ 3000 },
      "grafana": []int{ 3000 },
      "hivision": []int{ 7088 },
      "ejinshan": []int{ 6868 },
      "seeyon": []int{ 8001 },
      "java_ws": []int{ 8887 },
      "ifw8": []int{ 880 },
      "zabbix": []int{ 8069,10050 },
      "nacos": []int{ 7848,8848,9848,9849 },
      "mail": []int{ 25,465,587,2525,109,110,995,143,993 },
      "pop2": []int{ 109 },
      "pop3": []int{ 110,995 },
      "imap": []int{ 143,993 },
      "smtp": []int{ 25,465,587,2525 },
      "database1": []int{ 1433,1434,210,1158,1521,3306,3307,3308,5432,6379,63790,11211,27017,28017,7687 },
      "database2": []int{ 1433,1434,210,1158,1521,3306,3307,3308,4100,5000,5432,5984,6379,63790,11211,16000,16010,16020,16030,27017,28017,9001,9042,9160,54321,5236,7687 },
      "mysql": []int{ 3306,3307,3308 },
      "mssql": []int{ 1433,1434 },
      "oracle": []int{ 210,1158,1521 },
      "hsqldb": []int{ 9001 },
      "redis": []int{ 6379,63790 },
      "postgresql": []int{ 5432 },
      "mongodb": []int{ 27017,28017 },
      "db2": []int{ 5000 },
      "sybase": []int{ 4100,5000 },
      "couchdb": []int{ 5984 },
      "memcache": []int{ 11211 },
      "hbase": []int{ 16000,16010,16020,16030 },
      "cassandra": []int{ 9042,9160 },
      "kingbase8": []int{ 54321 },
      "dameng": []int{ 5236 },
      "neo4j": []int{ 7687 },
      "win": []int{ 22,2222,21,115,2121,23,88,135,593,5800,5900,5901,139,389,636,3268,3269,445,1080,3389,5985,5986,123 },
      "linux": []int{ 22,2222,21,115,2121,23,512,513,514,5800,5900,5901,6000,2049,43,1080,123,500,873,111,623,1026 },
      "mac": []int{ 22,2222,548,5800,5900,5901,2049 },
      "iiot": []int{ 20000,502,102,44818,1962,10001,5007,9600,789,1200,2404,20547 },
      "dnp": []int{ 20000 },
      "modbus": []int{ 502 },
      "s7": []int{ 102 },
      "ethernet": []int{ 44818 },
      "pcworx": []int{ 1962 },
      "atg": []int{ 10001 },
      "melsecq": []int{ 5007 },
      "omron": []int{ 9600 },
      "crimson": []int{ 789 },
      "codesys": []int{ 1200 },
      "iec104": []int{ 2404 },
      "procon": []int{ 20547 },
      "kerberos": []int{ 88 },
      "netbios": []int{ 139 },
      "smb": []int{ 139,445 },
      "rdp": []int{ 3389 },
      "winrm": []int{ 5985,5986 },
      "afp": []int{ 548 },
      "ftp": []int{ 21,115,2121 },
      "whois": []int{ 43 },
      "dns": []int{ 53 },
      "socks": []int{ 1080 },
      "oracle_ftp": []int{ 2100 },
      "ssh": []int{ 22,2222 },
      "ntp": []int{ 123 },
      "isakmp": []int{ 500 },
      "printer": []int{ 9100 },
      "mqtt": []int{ 1883 },
      "ajp": []int{ 8009 },
      "vnc": []int{ 5800,5900,5901 },
      "rsync": []int{ 873 },
      "nfs": []int{ 2049 },
      "sangfor": []int{ 51111 },
      "nodejs_debug": []int{ 5858,9229 },
      "telnet": []int{ 23 },
      "rpc": []int{ 111 },
      "msrpc": []int{ 135,593 },
      "irc": []int{ 194,6660 },
      "ldap": []int{ 389,636,3268,3269 },
      "rtsp": []int{ 554,8554 },
      "ipmi": []int{ 623 },
      "rusersd": []int{ 1026 },
      "amqp": []int{ 5672 },
      "kafka": []int{ 9092 },
      "upnp": []int{ 49152 },
      "hp": []int{ 5555,5556 },
      "altassian": []int{ 4990 },
      "lotus": []int{ 1352 },
      "cisco": []int{ 4786 },
      "lpd": []int{ 515 },
      "php_xdebug": []int{ 9000 },
      "hashicorp": []int{ 8500 },
      "checkpoint": []int{ 264 },
      "pcanywhere": []int{ 5632 },
      "docker": []int{ 2375,2376,2377,8080,8081,10252,6443,10256,31442,4149,10248,10250,10255,6781,6782,6783,2379,2380 },
      "docker_api": []int{ 2375,2376,2377 },
      "kubectl_manager": []int{ 10252 },
      "kubectl_proxy": []int{ 8080,8081 },
      "kube_apiserver": []int{ 6443,8080 },
      "kube_proxy": []int{ 10256,31442 },
      "kubelet_api": []int{ 4149,10248,10250,10255 },
      "kube_weave": []int{ 6781,6782,6783 },
      "kubeflow_dashboard": []int{ 8080 },
      "etcd": []int{ 2379,2380 },
      "iscsi": []int{ 3260 },
      "saprouter": []int{ 3299 },
      "distcc": []int{ 3632 },
      "zoho": []int{ 8383 },
      "phone": []int{ 46888 },
      "svn": []int{ 3690 },
      "snmp": []int{ 161 },
      "epmd": []int{ 4369 },
      "hadoop": []int{ 8020,8040,8041,8042,8088,8480,8485,9000,9083,10000,10003,14000,19888,41414,50010,50020,50030,50060,50070,50075,50090,50470,50475,60010,60030 },
      "rmi": []int{ 1098,4444,4445,8083,1028,1090,11099,47001,10999,1099 },
      "jndi": []int{ 1098,4444,4445,8083,1028,1090,11099,47001,10999,1099,1000,1001,1100,1101,5001,9999,10001,19001 },
      "jmx": []int{ 8093,8686,9010,9011,9012,50500,61616 },
      "jdwp": []int{ 3999,5000,5005,8000,8080,8453,8787,8788,9001,12001,12002,18000,45000,45001 },
      "rlogin": []int{ 512,513,514 },
      "glassfish": []int{ 4848 },
      "rocketmq": []int{ 9876,10909,10911,10912 },
      "vmware": []int{ 9875,5480 },
      "x11": []int{ 6000 },
      "legendsec": []int{ 48620 },
      "log4j": []int{ 4712 },
    }
    portGroupMap   = make(map[int][]string)
    portServersMap = make(map[string]string)
    rawCommonPorts = "in"
    commonPorts    = ParsePortRange(rawCommonPorts, false)
    commonPortsMap = GetObjectMap(commonPorts)
)

func usage() {
    fmt.Fprintf(
        os.Stdout, `
                          ...                                     .
                        .111111111111111.........................1111
      ......111..    .10011111011111110000000000000000111111111100000
  10010000000011.1110000001.111.111......1111111111111111..........
  10twelve0111...   .10001. ..
  100011...          1001               MX1014 by L
  .001              1001               Version 2.5.0
  .1.              ...1.


Usage: ./mx1014 [Options] [Target1] [Target2]...

Target Example:
    192.168.1.0/24
    192.168.1.*
    192.168.1-12.1
    192.168.*.1:22,80-90,8080
    github.com:22,443,rce

Options:
`)
    flagSet := flag.CommandLine
    options := map[string][]string{
        "Target":  []string{"i", "I", "g", "sh", "cnet", "r", "R"},
        "Port":    []string{"p", "sp", "ep", "hp", "fuzz"},
        "Connect": []string{"t", "T", "u", "e", "A", "a", "proxy"},
        "Output":  []string{"o", "c", "d", "D", "l", "P", "v"},
    }
    for _, category := range []string{"Target", "Port", "Connect", "Output"} {
        fmt.Printf("  [%s]\n", category)
        for _, name := range options[category] {
            fl4g := flagSet.Lookup(name)
            fmt.Printf("    -%s", fl4g.Name)
            fmt.Printf(" %s\n", fl4g.Usage)
        }
        fmt.Print("\n")
    }
    os.Exit(0)
}

func init() {
    // Target
    flag.StringVar(&infile, "i", "", " File   Target input from list")
    flag.BoolVar(&ignoreErrHost, "I", false, "        Ignore the wrong address and continue scanning")
    flag.StringVar(&gatewayRanges, "g", "", " Net    Intranet gateway address range (10/172/192/all)")
    flag.BoolVar(&showHosts, "sh", false, "       Show scan target")
    flag.BoolVar(&cNet, "cnet", false, "     C net mode")
    flag.BoolVar(&rejectAllOpen, "r", false, "        Reject all open targets")
    flag.IntVar(&rejectAllOpenTimes, "R", 1, " Int    Reject all open of tested (Default is 1)")

    // Port
    flag.StringVar(&portRanges, "p", rawCommonPorts, " Ports  Default port ranges (Default is \"in\" port group)")
    flag.BoolVar(&showPorts, "sp", false, "       Only show default ports (see -p)")
    flag.StringVar(&excludePortRanges, "ep", "", "Ports  Exclude port (see -p)")
    flag.StringVar(&headPortRanges, "hp", "80,443,8080,22,445,3389,in", "Ports  Priority scan port (Default 80,443,8080,22,445,3389,in)")
    flag.BoolVar(&fuzzPort, "fuzz", false, "     Fuzz Port")

    // Connect
    flag.IntVar(&numOfgoroutine, "t", 512, " Int    The Number of Goroutine (Default is 512)")
    flag.IntVar(&timeout, "T", 1980, " Int    TCP Connect Timeout (Default is 1980ms)")
    flag.BoolVar(&udpmode, "u", false, "        UDP spray")
    flag.BoolVar(&echoMode, "e", false, "        Echo mode (TCP needs to be manually)")
    flag.BoolVar(&forceScan, "A", false, "        Disable auto discard")
    flag.IntVar(&autoDiscard, "a", 512, " Int    Too many filtered, Discard the host (Default is 512)")
    flag.StringVar(&proxy, "proxy", "", " Str    SOCKS5 proxy address (socks5://user:pass@host:port)")

    // Output
    flag.StringVar(&outfile, "o", "", " File   Output file path")
    flag.BoolVar(&closedMode, "c", false, "        Allow display of closed ports (Only TCP)")
    flag.StringVar(&senddata, "d", "%port%\n", " Str    Specify Echo mode data (Default is \"%port%\\n\")")
    flag.IntVar(&progressDelay, "D", 7, " Int    Progress Bar Refresh Delay (Default is 7s)")
    flag.BoolVar(&aliveMode, "l", false, "        Output alive host")
    flag.BoolVar(&disableProtocolName, "P", false, "        Do not output protocol name")
    flag.BoolVar(&verbose, "v", false, "        Verbose mode")
    flag.Usage = usage

    // initialize the port map
    for name, ports := range portGroup {
        for _, port := range ports {
            if portGroupMap[port] == nil {
                portGroupMap[port] = []string{}
            }
            portGroupMap[port] = append(portGroupMap[port], name)
        }
    }
    for port, servers := range portGroupMap {
        portServersMap[strconv.Itoa(port)] = strings.Join(servers, ",")
    }
}

func Run() {

    SetUlimit()

    flag.Parse()
    log.SetFlags(0)
    if outfile != "" {
        logFile, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModeAppend|os.ModePerm)
        if err != nil {
            ErrPrint("Open output file failed")
        }

        defer logFile.Close()
        out := io.MultiWriter(os.Stdout, logFile)
        log.SetOutput(out)
    } else {
        out := io.MultiWriter(os.Stdout)
        log.SetOutput(out)
    }

    if proxy != "" {
        if !strings.HasPrefix(proxy, "socks5://") {
            ErrPrint("Invalid proxy scheme, must be socks5://")
        }
        rest := proxy[9:]
        atIndex := strings.LastIndex(rest, "@")
        if atIndex != -1 {
            userPass := rest[:atIndex]
            proxyAddr = rest[atIndex+1:]
            colonIndex := strings.Index(userPass, ":")
            if colonIndex != -1 {
                proxyAuth = &socks5Auth{username: userPass[:colonIndex], password: userPass[colonIndex+1:]}
            } else {
                proxyAuth = &socks5Auth{username: userPass}
            }
        } else {
            proxyAddr = rest
        }
        if proxyAddr == "" {
            ErrPrint("Invalid proxy address")
        }
        if udpmode {
            ErrPrint("UDP mode (-u) is not supported with SOCKS5 proxy (-proxy)")
        }
        log.Printf("# Using SOCKS5 proxy: %s\n", proxyAddr)
    }

    defaultPorts := ParsePortRange(portRanges, false)
    defaultPortsLen = uint64(len(defaultPorts))
    if showPorts {
        fmt.Printf("# Count: %d\n", defaultPortsLen)
        fmt.Println(strings.Join(defaultPorts, ","))
        os.Exit(0)
    }

    // parse target
    var rawTargets []string
    rawTargets = flag.Args()

    if infile != "" {
        rawTargets = append(rawTargets, FileReadlines(infile)...)
    }

    if cNet {
        var newRawTargets []string
        for _, rawTarget := range rawTargets {
            cidr := rawTarget + "/24"
            _, ipnet, err := net.ParseCIDR(cidr)
            if err == nil {
                newRawTargets = append(newRawTargets, ipnet.String())
            } else {
                newRawTargets = append(newRawTargets, rawTarget)
            }
        }
        rawTargets = newRawTargets
    }

    if gatewayRanges != "" {
        if gatewayRanges == "all" {
            gatewayRanges = "10,172,192"
        }
        for _, gatewayNet := range strings.Split(gatewayRanges, ",") {
            switch gatewayNet {
            case "10":
                rawTargets = append(rawTargets, "10.*.*.1", "10.*.*.254")
            case "172":
                rawTargets = append(rawTargets, "172.16-31.*.1", "172.16-31.*.254")
            case "192":
                rawTargets = append(rawTargets, "192.168.*.1", "192.168.*.254")
            default:
                ErrPrint(fmt.Sprintf("Wrong gateway name (-g): %s", gatewayNet))
            }
        }
    }

    wg := sync.WaitGroup{}
    rawtargetChan := make(chan string, timeout)
    for i := 0; i <= numOfgoroutine; i++ {
        go func() {
            for rawTarget := range rawtargetChan {
                err := ParseTarget(rawTarget, defaultPorts)
                mutex.Lock()
                if err != nil {
                    if ignoreErrHost {
                        log.Printf("# Wrong target: %s", rawTarget)
                    } else {
                        ErrPrint(fmt.Sprintf("Wrong target: %s", rawTarget))
                    }

                }
                mutex.Unlock()
                wg.Done()
            }
        }()
    }
    for _, rawTarget := range RemoveRepeatedElement(rawTargets) {
        rawtargetChan <- rawTarget
        wg.Add(1)
    }
    wg.Wait()

    // exclude ports
    if excludePortRanges != "" {
        excludePorts := ParsePortRange(excludePortRanges, false)
        for _, eport := range excludePorts {
            if portMap[eport] != nil {
                for _, rawTarget := range portMap[eport] {
                    total -= uint64(len(hostMap[rawTarget]))
                }
                delete(portMap, eport)
            }
        }
    }

    if showHosts {
        fmt.Printf("# Count: %d\n", hostTotal)
        for _, hosts := range hostMap {
            fmt.Println(strings.Join(hosts, "\n"))
        }
        os.Exit(0)
    }

    if hostTotal == 0 {
        flag.Usage()
    }

    if rejectAllOpen {
        log.Printf("# %s Start automatically reject all-open targets, scanning %d hosts... (reqs: %d)\n", startTime.Format("2006/01/02 15:04:05"), hostTotal, hostTotal*uint64(rejectAllOpenTimes))
        RejectAllOpenTargets()
        endTime := time.Now().Format("2006/01/02 15:04:05")
        log.Printf("# %s Finished. reject all-open %d hosts.\n\n", endTime, rejectCount)
    }

    EchoModePrompt := ""
    if echoMode && !udpmode {
        EchoModePrompt = " (TCP Echo)"
    }
    if udpmode {
        EchoModePrompt = " (UDP Spray)"
    }
    log.Printf("# %s Start scanning %d hosts...%s (reqs: %d)\n\n", startTime.Format("2006/01/02 15:04:05"), hostTotal, EchoModePrompt, total)
    PortScan()
    spendTime := time.Since(startTime).Seconds()
    pps := uint64(float64(total) / spendTime)
    if pps > total {
        pps = total
    }
    aliveRate := uint64(hostUpCount) * 100.0 / hostTotal
    endTime := time.Now().Format("2006/01/02 15:04:05")
    log.Printf("\n# %s Finished %d tasks.\n", endTime, total)
    log.Printf("# up: %d%% (%d/%d), discard: %d, open: %d, pps: %d, time: %s\n", aliveRate, hostUpCount, hostTotal, hostDiscard, openCount, pps, secondToTime(uint64(spendTime)))
    if outfile != "" {
        log.Printf("# Save the result to \"%s\"\n", outfile)
    }
}
