package main

import (
	"fmt"
	"time"
	"context"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)
func getunixtime() int64 {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
			panic(err)
	}
	now := time.Now()
	t := now.In(loc)
	a := t.UnixNano() / 1000000
	return a
}

func processInfo() ([]string, error) {
    subs := []string{}

    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10) * time.Second)
    defer cancel()

    // map을 사용해서 프로세스 이름을 기준으로 CPU 사용량과 메모리 사용량을 모으기
    cpuMap := make(map[string]float64)
    memMap := make(map[string]float64)

    // 프로세스의 context(PID, CPU, MEM...)를 리턴하는 함수 ProcessesWithContext
    plist, err := process.ProcessesWithContext(ctx)

    if err != nil {
        fmt.Println("Failed to get processes:", err)
        return subs, err
    }

    // 시스템의 전체 메모리 용량 확인
    vmStat, err := mem.VirtualMemory()
    if err != nil {
        panic(err)
    }
    totalMem := float64(vmStat.Total)

    // 각 프로세스의 CPU, MEM 사용률 체크하기
    for _, proc := range plist {
        name, err := proc.Name()
        if err != nil {
            fmt.Print("Failed to get process name:", err)
        }
        // CPU 사용률 체크
        cpuPercent, err := proc.CPUPercent()
        if err != nil {
            continue
        }
        cpuMap[name] += cpuPercent

        // MEM 사용률 체크
        memInfo, err := proc.MemoryInfo()
        if err != nil {
            continue
        }
        memMap[name] += float64(memInfo.RSS) / totalMem * 100.0
    }


    // map에 저장된 정보를 InfluxDB 형식으로 출력하기
    nowtime := getunixtime()
    for name, cpuPercent := range cpuMap {
        memPercent, ok := memMap[name]
        if !ok {
            // 해당 프로세스의 메모리 사용량 정보가 없으면 건너뜀
            continue
        }
        fmt.Printf("pro_cpu_percent{name=\"%s\"} %.3f %d\n", name, cpuPercent, nowtime)
        fmt.Printf("pro_mem_percent{name=\"%s\"} %.3f %d\n", name, memPercent, nowtime)
    }

    return subs, nil
}


func main() {
	processInfo()
}
