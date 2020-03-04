package main

import (
  "os"
  "fmt"
  "bufio"
  "strconv"
  "strings"
  node_client "../robot-communication-framework/rcf_cc_node_client"
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
      if len(cmd_args) >=1 {
        node_client.Create_topic(conn, cmd_args[1])
      }
    } else if string(cmd_args[0]) == "cpulld" {
      if len(cmd_args) >=1 {
        topic_listener := node_client.Continuous_data_pull(conn, cmd_args[1])
        for {
          select {
            case data := <-topic_listener:
              fmt.Println("data changed", data)
          }
        }
      }
    } else if string(cmd_args[0]) == "pushd" {
      if len(cmd_args) >=2 {
        node_client.Push_data(conn, cmd_args[1], cmd_args[2])
      }
    } else if string(cmd_args[0]) == "pulld" {
      if len(cmd_args) >=2 {
        nele,_ := strconv.Atoi(cmd_args[2])
        elements := node_client.Pull_data(conn, nele, cmd_args[1])
        fmt.Println(elements)
      }
    } else if string(cmd_args[0]) == "lt" {
        if len(cmd_args) >=0 {
          topic_names := node_client.List_cctopics(conn)
          fmt.Println(topic_names)
        }
    } else if string(cmd_args[0]) == "end" {
        if len(cmd_args) >=0 {
          node_client.Close_cc_node(conn)
          return
        }
    }
  }

  node_client.Close_cc_node(conn)
}
