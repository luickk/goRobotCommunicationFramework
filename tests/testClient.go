package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"

	"rcf/rcfNodeClient"
)

func main() {
  client, err := rcfNodeClient.New(47)

	if err != nil {
		fmt.Println(err)
		return
	}

  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter command: ")
    cmd_txt,_ := reader.ReadString('\n')
    cmd_txt = strings.Replace(cmd_txt, "\n", "", -1)
    cmd_args :=strings.Split(cmd_txt, " ")

    if string(cmd_args[0]) == "ct" {
      if len(cmd_args) >=1 {
        client.TopicCreate(cmd_args[1])
      }
    }  else if string(cmd_args[0]) == "sub" {
      if len(cmd_args) >=1 {
        topic_listener := client.TopicDataSubscribe(cmd_args[1])
        for {
          select {
            case data := <-topic_listener:
              fmt.Println("data changed: ", string(data))
          }
        }
      }
    } else if string(cmd_args[0]) == "pushd" {
      if len(cmd_args) >=2 {
        client.TopicPublishData(cmd_args[1], []byte(cmd_args[2]))
      }
    } else if string(cmd_args[0]) == "pulld" {
      if len(cmd_args) >=2 {
        nele,_ := strconv.Atoi(cmd_args[2])
        elements := client.TopicPullData(cmd_args[1], nele)
				fmt.Println(elements)
      }
    } else if string(cmd_args[0]) == "lt" {
      if len(cmd_args) >=0 {
        topicNames := client.TopicList()
        fmt.Println(topicNames)
      }
    } else if string(cmd_args[0]) == "ea" {
      if len(cmd_args) >=1 {
        client.ActionExec(cmd_args[1], []byte("testparam"))
      }
    } else if string(cmd_args[0]) == "es" {
      if len(cmd_args) >=1 {
        result := client.ServiceExec(cmd_args[1], []byte("testparam"))
        fmt.Println(string(result))
      }
    }
  }


}
