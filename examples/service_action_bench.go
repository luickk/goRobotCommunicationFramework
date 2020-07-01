package main

import (
	rcfNodeClient "rcf/rcfNodeClient"
	"time"
	"strconv"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	Nexeced := 0

	go func() {
		for {
			time.Sleep(1 * time.Second)

			println("Benchmark, execed services/ actions in one second: " + strconv.Itoa(Nexeced))

			Nexeced = 0
		}
	}()

	// smaple loop
	for {
		time.Sleep(1000*time.Microsecond)
		
		Nexeced++
	
		go func() {
			res := rcfNodeClient.ServiceExec(client, "testService", []byte("randTestParamFromServiceBench"))
			println("results delay: " + string(res))
			Nexeced++
		}()
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
