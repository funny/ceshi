package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/1046102779/utest"
	"github.com/shirou/gopsutil/mem"
)

func memory(fileName string) (string, error) {
	var (
		output string
		err    error
		v      *mem.VirtualMemoryStat
	)

	if v, err = mem.VirtualMemory(); err != nil {
		return "", err
	}
	output = fmt.Sprintf("Total: %v, Free:%v, UsedPercent:%f%%\n",
		v.Total,
		v.Free,
		v.UsedPercent)
	return output, err
}

func main() {
	utest.RegisterCmdHandler("memory", memory)
	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			fmt.Printf("goroutine %d: running %ds\n", i+1, 1)
			wg.Done()
		}()
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
	return
}
