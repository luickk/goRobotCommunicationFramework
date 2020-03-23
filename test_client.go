package main

import (
  "os"
  "fmt"
  "bufio"
  "strconv"
  "strings"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  conn := node_client.Node_open_conn(28)

  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter command: ")
    cmd_txt,_ := reader.ReadString('\n')
    cmd_txt = strings.Replace(cmd_txt, "\n", "", -1)
    cmd_args :=strings.Split(cmd_txt, " ")

    if string(cmd_args[0]) == "ct" {
      if len(cmd_args) >=1 {
        node_client.Topic_create(conn, cmd_args[1])
      }
    } else if string(cmd_args[0]) == "cpulld" {
      if len(cmd_args) >=1 {
        topic_listener := node_client.Topic_subscribe(conn, cmd_args[1])
        for {
          select {
            case data := <-topic_listener:
              fmt.Println("data changed: ", data)
          }
        }
      }
    } else if string(cmd_args[0]) == "pushd" {
      if len(cmd_args) >=2 {
        data_map := make(map[string]string)
        data_map["cli"] = cmd_args[2]
        node_client.Topic_publish_data(conn, cmd_args[1], "data_map")
      }
    } else if string(cmd_args[0]) == "pulld" {
      if len(cmd_args) >=2 {
        nele,_ := strconv.Atoi(cmd_args[2])
        elements := node_client.Topic_pull_data(conn, nele, cmd_args[1])
        fmt.Println(elements)
      }
    } else if string(cmd_args[0]) == "lt" {
      if len(cmd_args) >=0 {
        topic_names := node_client.Topic_list(conn)
        fmt.Println(topic_names)
      }
    } else if string(cmd_args[0]) == "es" {
      if len(cmd_args) >=1 {
        node_client.Action_exec(conn, cmd_args[1])
      }
    } else if string(cmd_args[0]) == "end" {
        if len(cmd_args) >=0 {
          node_client.Node_close_conn(conn)
          return
        }
    }
  }

  node_client.Node_close_conn(conn)
}
