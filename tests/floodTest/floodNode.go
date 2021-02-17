package main

import (
	"fmt"
	"time"

	"goRobotCommunicationFramework/rcfNode"
)

func main() {
	errorStream := make(chan error)
  node, err := rcfNodeClient.New(8000, errorStream)
	if err != nil {
		fmt.Println(err)
		return
	}

  node.ActionCreate("testAction", func(params []byte, n rcfNode.Node){
    fmt.Println("- test Action executed" + string(params))
  })

  node.ServiceCreate("testService", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("- test Service executed: " + string(params))
    return []byte("result")
  })

  node.ServiceCreate("testDelayService", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("- test delay Service executed: " + string(params))
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  for{}
}
