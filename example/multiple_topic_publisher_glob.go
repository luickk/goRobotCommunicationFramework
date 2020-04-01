package main

import (
  "fmt"
  "time"
  "math/rand"
  "strconv"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  conn := node_client.Node_open_conn(30)

  // creating topic by sending cmd to node
  node_client.Topic_create(conn, "altsens")
  node_client.Topic_create(conn, "radarsens")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    random := rand.Intn(100)
    // printing sample data
    fmt.Println(random)
    // putting sample data into map
    rad_data_map := make(map[string]string)
    rad_data_map["rad"] = strconv.Itoa(random)

    alt_data_map := make(map[string]string)
    alt_data_map["alt"] = strconv.Itoa(random)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    node_client.Topic_glob_publish_data(conn, "radarsens", rad_data_map)
    node_client.Topic_glob_publish_data(conn, "altsens", alt_data_map)

    time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  node_client.Node_close_conn(conn)
}
