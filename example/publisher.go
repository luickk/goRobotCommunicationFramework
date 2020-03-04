package main

import (
  "fmt"
  "time"
  "math/rand"
  "strconv"
  node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {
  conn := node_client.Connect_to_cc_node(30)

  node_client.Create_topic(conn, "altsens")

  for {
    time.Sleep(1 * time.Second)
    alt := rand.Intn(100)
    fmt.Println(alt)
    node_client.Push_data(conn, "altsens", strconv.Itoa(alt))
  }

  node_client.Close_cc_node(conn)
}
