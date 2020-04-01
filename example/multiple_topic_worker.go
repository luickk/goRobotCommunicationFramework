package main

import (
  "fmt"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  conn := node_client.Node_open_conn(30)

  // initiating topic listener
  // returns channel which every new incoming element/ msg is pushed to
  alt_topic_listener := node_client.Topic_subscribe(conn, "altsens")
  rad_topic_listener := node_client.Topic_subscribe(conn, "radarsens")

  // smaple loop
  for {
    // select statement to wait for new incoming elements/msgs from listened to topic
    select {
      // if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
    case msg := <-alt_topic_listener:
          // printing new altitude, pushed to topic
          fmt.Println("Altitude changed: ", msg)
    case msg := <-rad_topic_listener:
          // printing new altitude, pushed to topic
          fmt.Println("Radar changed: ", msg)
    }
  }


  // closing node conn at program end
  node_client.Node_close_conn(conn)
}
