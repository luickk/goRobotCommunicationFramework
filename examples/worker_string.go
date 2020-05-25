package main

import (
	"fmt"
	"strconv"
	"strings"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  conn := nodeClient.NodeOpenConn(47)

  // initiating topic listener
  // returns channel which every new incoming element/ msg is pushed to
  topicListener := nodeClient.TopicStringDataSubscribe(conn, "altsensstring")

  // smaple loop
  for {
    // select statement to wait for new incoming elements/msgs from listened to topic
    select {
      // if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
      case alt := <-topicListener:
          // converting altitude element/ msg which is encoded as string to integer
          // removing spaces before
          alti,_ := strconv.Atoi(strings.TrimSpace(alt))
          // printing new altitude, pushed to topic
          fmt.Println("Altitude string changed: ", alti)
          // checking if new altitude is greater than 90 for example purposes
          if alti >= 90 {
            // printing action call alert
            fmt.Println("exec service")
            // executing service "testService" on connected node
            // service must be initiated/ provided by the node
            nodeClient.ServiceExec(conn, "testService", []byte(""))
          }
    }
  }


  // closing node conn at program end
  nodeClient.NodeCloseConn(conn)
}
