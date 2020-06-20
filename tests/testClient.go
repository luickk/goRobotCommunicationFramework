package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
  client := rcfNodeClient.NodeOpenConn(47)

  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter command: ")
    cmd_txt,_ := reader.ReadString('\n')
    cmd_txt = strings.Replace(cmd_txt, "\n", "", -1)
    cmd_args :=strings.Split(cmd_txt, " ")

    if string(cmd_args[0]) == "ct" {
      if len(cmd_args) >=1 {
        rcfNodeClient.TopicCreate(client, cmd_args[1])
      }
    } else if string(cmd_args[0]) == "gsub" {
      if len(cmd_args) >=1 {
        topic_listener := rcfNodeClient.TopicGlobDataSubscribe(client, cmd_args[1])
        for {
          select {
            case data := <-topic_listener:
              fmt.Println("data changed: ", data)
          }
        }
      }
    } else if string(cmd_args[0]) == "gpushd" {
      if len(cmd_args) >=2 {
        data_map := make(map[string]string)
        data_map["cli"] = cmd_args[2]
        rcfNodeClient.TopicPublishGlobData(client, cmd_args[1], data_map)
      }
    } else if string(cmd_args[0]) == "gpulld" {
      if len(cmd_args) >=2 {
        nele,_ := strconv.Atoi(cmd_args[2])
        elements := rcfNodeClient.TopicPullGlobData(client, nele, cmd_args[1])
        fmt.Println(elements)
      }
    }  else if string(cmd_args[0]) == "sub" {
      if len(cmd_args) >=1 {
        topic_listener := rcfNodeClient.TopicStringDataSubscribe(client, cmd_args[1])
        for {
          select {
            case data := <-topic_listener:
              fmt.Println("data changed: ", data)
          }
        }
      }
    } else if string(cmd_args[0]) == "pushd" {
      if len(cmd_args) >=2 {
        rcfNodeClient.TopicPublishStringData(client, cmd_args[1], cmd_args[2])
      }
    } else if string(cmd_args[0]) == "pulld" {
      if len(cmd_args) >=2 {
        nele,_ := strconv.Atoi(cmd_args[2])
        elements := rcfNodeClient.TopicPullStringData(client, nele, cmd_args[1])
        fmt.Println(elements)
      }
    } else if string(cmd_args[0]) == "lt" {
      if len(cmd_args) >=0 {
        topicNames := rcfNodeClient.TopicList(client)
        fmt.Println(topicNames)
      }
    } else if string(cmd_args[0]) == "ea" {
      if len(cmd_args) >=1 {
        rcfNodeClient.ActionExec(client, cmd_args[1], []byte("testparam"))
      }
    } else if string(cmd_args[0]) == "es" {
      if len(cmd_args) >=1 {
        result := rcfNodeClient.ServiceExec(client, cmd_args[1], []byte("testparam"))
        fmt.Println(string(result))
      }
    } else if string(cmd_args[0]) == "end" {
        if len(cmd_args) >=0 {
          rcfNodeClient.NodeCloseConn(client)
          return
        }
    }
  }

  rcfNodeClient.NodeCloseConn(client)
}
