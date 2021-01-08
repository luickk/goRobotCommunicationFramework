package main

import (
	"fmt"
	"time"

	"rcf/rcfNode"
)

func main() {
  node, err := rcfNode.New(47)
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
