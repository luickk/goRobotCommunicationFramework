package main

import (
  "os"
  "fmt"
  "bufio"
  "strings"
  node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {
  conn := node_client.Connect_to_cc_node(28)

  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter text: ")
    cmd_txt,_ := reader.ReadString('\n')
    cmd_txt = strings.Replace(cmd_txt, "\n", "", -1)
    cmd_args :=strings.Split(cmd_txt, " ")

    if string(cmd_args[0]) == "ct" {
      node_client.Create_topic(conn, cmd_args[1])
    } else if string(cmd_args[0]) == "cp" {
      node_client.Push_data(conn, cmd_args[1], cmd_args[2])
    } else if string(cmd_args[0]) == "end" {
      node_client.Close_cc_node(conn)
      return
    }
  }

  // node_client.Push_data(conn, "b1", "test")
  // node_client.Push_data(conn, "b1", "test")

  // fmt.Println(node_client.Pull_data(conn, 3, "test"))

  node_client.Close_cc_node(conn)
}
