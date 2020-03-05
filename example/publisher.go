package main

import (
  "fmt"
  "time"
  "math/rand"
  "strconv"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  conn := node_client.Node_open_conn(30)

  node_client.Topic_create(conn, "altsens")

  for {
    time.Sleep(1 * time.Second)
    alt := rand.Intn(100)
    fmt.Println(alt)
    node_client.Topic_push_data(conn, "altsens", strconv.Itoa(alt))
  }

  node_client.Node_close_conn(conn)
}
