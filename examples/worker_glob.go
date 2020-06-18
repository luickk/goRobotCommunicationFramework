package main

import (
	"fmt"
  "strconv"
  "math/rand"
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  client := rcfNodeClient.NodeOpenConn(47)

  // initiating topic listener
  // returns channel which every new incoming element/ msg is pushed to
  topicListener := rcfNodeClient.TopicGlobDataSubscribe(client, "altsensglob")

  // smaple loop
  for {
    // select statement to wait for new incoming elements/msgs from listened to topic
    select {
      // if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
    case globData := <-topicListener:
      // converting altitude element/ msg which is encoded as string to integer
      // removing spaces before
      alti,_ := strconv.Atoi(globData["alt"])
      // printing new altitude, pushed to topic
      fmt.Println("Altitude glob changed: ", alti)
      // checking if new altitude is greater than 90 for example purposes
      if alti == 99 {
        rand := strconv.Itoa(rand.Intn(1000000))
        // printing action call alert
        // fmt.Println("called action")
        // // calling action "test" on connected node
        // // action must be initiated/ provided by the node
        rcfNodeClient.ActionExec(client, "test", []byte(""))
        println("exec service: "+rand)
        
        go func() { 
          res := rcfNodeClient.ServiceExec(client, "testServiceDelay", []byte("randTestParamFromGlobWorker"+rand))
          println("results delay: "+string(res))
        }()
      }
    }
  }


  // closing node conn at program end
  rcfNodeClient.NodeCloseConn(client)
}
