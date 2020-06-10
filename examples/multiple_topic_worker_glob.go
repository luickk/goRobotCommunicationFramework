package main

import (
	"fmt"
  "strconv"
  "math/rand"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 47
  client := nodeClient.NodeOpenConn(47)

  // initiating topic listener
  // returns channel which every new incoming element/ msg is pushed to
  altTopicListener := nodeClient.TopicGlobDataSubscribe(client, "altsensmglob")
  radTopicListener := nodeClient.TopicGlobDataSubscribe(client, "radarsensmglob")

  // smaple loop
  for {
    // select statement to wait for new incoming elements/msgs from listened to topic
    select {
        // if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
      case msg := <-altTopicListener:
        // converting altitude element/ msg which is encoded as string to integer
        // removing spaces before
        alti,_ := strconv.Atoi(msg["alt"])
        // printing new altitude, pushed to topic
        fmt.Println("Altitude multi glob changed: ", alti)
        // cchecking if new altitude is greater than 90 for example purposes
        if alti >= 90 {
          // printing action call alert
          fmt.Println("called action")
          // calling action "testAction" on connected node
          // action must be initiated/ provided by the node
          nodeClient.ActionExec(client, "testAction", []byte(""))
        }
      case msg := <-radTopicListener:
        // converting altitude element/ msg which is encoded as string to integer
        // removing spaces before
        rad,_ := strconv.Atoi(msg["rad"])

        rand := strconv.Itoa(rand.Intn(1000000))
        
        // printing new altitude, pushed to topic
        fmt.Println("Radar multi glob changed: ", rad)
        if rad == 99 {
          // printing action call alert
          fmt.Println("exec service: " + rand)
          // executing service "testService" on connected node
          // service must be initiated/ provided by the node
          // res := nodeClient.ServiceExec(client, "testService", []byte("testParamFromMultiTopicWorker"+rand))
          // println("results: " + string(res))
      }
    }
  }


  // closing node conn at program end
  nodeClient.NodeCloseConn(client)
}
