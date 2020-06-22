package main

import (
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	// smaple loop
	for {
		rcfNodeClient.ActionExec(client, "test", []byte(""))
		// go func() {
		// 	res := rcfNodeClient.ServiceExec(client, "testServiceDelay", []byte("randTestParamFromServiceBench"))
		// 	println("results delay: " + string(res))
		// }()

	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
