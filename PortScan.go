package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 获取用户输入
func IpAndPort() (string, []int) {
	//读取用户控制台输入
	reader := bufio.NewReader(os.Stdin)

	//获取目标主机
	fmt.Print("ip address: ")
	ip, err := reader.ReadString('\n') //当遇到换行符或者回车的时候停止读取
	if err != nil {
		fmt.Printf("IP is failed...\n", err)
		os.Exit(1)
	}
	ip = strings.TrimSpace(ip) //去掉首尾的特殊字符和空白字符
	if ip == "" {
		fmt.Printf("IP can not be empty...\n")
		os.Exit(1)
	}

	//获取端口
	fmt.Print("port: ")
	port, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Port is failed...\n", err)
		os.Exit(1)
	}
	port = strings.TrimSpace(port)
	if port == "" {
		fmt.Printf("Port can not be empty...\n")
		os.Exit(1)
	}

	//解析端口范围
	ports, err := AnalyzePort(port)
	if err != nil {
		fmt.Printf("Port is failed...\n", err)
		os.Exit(1)
	}
	if len(ports) == 0 {
		fmt.Printf("Port must have a port...\n")
		os.Exit(1)
	}

	return ip, ports
}

// 解析端口范围
func AnalyzePort(newPorts string) ([]int, error) {
	var ports []int
	//处理单个输入的端口
	if !strings.Contains(newPorts, ",") && !strings.Contains(newPorts, "-") {
		onePort, err := strconv.Atoi(newPorts)
		if err != nil {
			fmt.Printf("Port is failed...\n", err)
		}
		if onePort >= 1 && onePort <= 65535 {
			ports = append(ports, onePort)
		}
	} else if strings.Contains(newPorts, ",") && !strings.Contains(newPorts, "-") { //处理多个端口，且用,隔开

		anyports := strings.Split(newPorts, ",")
		for _, anyport := range anyports {
			anyPort, err := strconv.Atoi(anyport)
			if err != nil {
				fmt.Println(err)
			}
			ports = append(ports, anyPort)
		}
	} else if strings.Contains(newPorts, "-") && !strings.Contains(newPorts, ",") { //处理范围端口,用-隔开
		hengPorts := strings.Split(newPorts, "-")
		fmt.Println(hengPorts)
		hengPortstart, _ := strconv.Atoi(hengPorts[0])
		hengPortend, _ := strconv.Atoi(hengPorts[1])
		if hengPortstart >= 1 && hengPortend <= 65535 && hengPortstart < hengPortend {
			for i := hengPortstart; i <= hengPortend; i++ {
				ports = append(ports, i)
			}
		} else {
			fmt.Printf("Port can not start more than 65535 or less than 1...\n")
		}
	} else if strings.Contains(newPorts, ",") && strings.Contains(newPorts, "-") {
		kindPorts := strings.Split(newPorts, ",")
		for _, kindPort := range kindPorts {
			if strings.Contains(kindPort, "-") {
				kindPortHeng := strings.Split(kindPort, "-")
				kindPortStart, _ := strconv.Atoi(kindPortHeng[0])
				kindPortEnd, _ := strconv.Atoi(kindPortHeng[1])
				if kindPortStart >= 1 && kindPortEnd <= 65535 && kindPortStart < kindPortEnd {
					for i := kindPortStart; i <= kindPortEnd; i++ {
						ports = append(ports, i)
					}
				}
			}
			kindPortNormal, _ := strconv.Atoi(kindPort)
			ports = append(ports, kindPortNormal)
		}
	}
	return ports, nil
}

func PortScan(ip string, ports []int) []int {
	fmt.Printf("------------------------------ PortScan starting...------------------------------\n")

	var wg sync.WaitGroup           //等待一组goroutine完成任务
	openPorts := make([]int, 0)     //存储扫描到的开放端口
	portChan := make(chan int, 100) //创建一个带缓冲的通道
	var mu sync.Mutex               //创建一个互斥锁，保护共享资源

	//启动工作协程
	worker := 100
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portChan {
				if IsPortOpen(ip, port, 2*time.Second) {
					mu.Lock() //加锁
					openPorts = append(openPorts, port)
					mu.Unlock() //解锁
					fmt.Printf("Port %d is open\n", port)
				}
			}
		}()
	}
	//发送端口到通道
	go func() {
		for _, port := range ports {
			portChan <- port
		}
		close(portChan)
	}()
	//等待所有工作协程完成
	wg.Wait()

	return openPorts
}

// 检查端口是否开放
func IsPortOpen(ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// 输出扫描结果
func PrintPort(openports []int) {
	fmt.Printf("------------------------------ PortScan finish...------------------------------\n")
	fmt.Printf("Open ports: %d\n", len(openports))
	if len(openports) == 0 {
		fmt.Println("No open ports")
		for _, port := range openports {
			fmt.Printf("Port %d is open\n", port)
		}
	}
}

func main() {
	// 获取用户输入
	ip, ports := IpAndPort()

	// 扫描端口
	openPorts := PortScan(ip, ports)

	// 输出结果
	PrintPort(openPorts)
}
