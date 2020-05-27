package main

import (
	"fmt"
	"strconv"
	"strings"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  connChan, conn := nodeClient.NodeOpenConn(47)

  // initiating topic listener
  // returns channel which every new incoming element/ msg is pushed to
  topicListener := nodeClient.TopicStringDataSubscribe(conn, connChan, "altsensstring")

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
    }
  }


  // closing node conn at program end
  nodeClient.NodeCloseConn(conn)
}
