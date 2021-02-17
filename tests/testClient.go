package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"

	"goRobotCommunicationFramework/rcfNodeClient"
)

func main() {
	errorStream := make(chan error)
  client, err := rcfNodeClient.New(8000, errorStream)
	if err != nil {
		fmt.Println(err)
		return
	}

  reader := bufio.NewReader(os.Stdin)
  for {
		select {
		case err := <-errorStream:
			fmt.Println(err)
			return
		default:
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
	        topic_listener, err := client.TopicDataSubscribe(cmd_args[1])
					if err != nil {
						fmt.Println(err)
						return
					}
	        for {
	          select {
	            case data := <-topic_listener:
	              fmt.Println("data changed: ", string(data))
	          }
	        }
	      }
	    } else if string(cmd_args[0]) == "pushd" {
	      if len(cmd_args) >=2 {
	        if err := client.TopicPublishData(cmd_args[1], []byte(cmd_args[2])); err != nil {
						fmt.Println(err)
						return
					}
	      }
	    } else if string(cmd_args[0]) == "pulld" {
	      if len(cmd_args) >=2 {
	        nele,_ := strconv.Atoi(cmd_args[2])
	        elements, err := client.TopicPullData(cmd_args[1], nele)
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println(elements)
	      }
	    } else if string(cmd_args[0]) == "lt" {
	      if len(cmd_args) >=0 {
	        topicNames, err := client.TopicList()
					if err != nil {
						fmt.Println(err)
						return
					}
	        fmt.Println(topicNames)
	      }
	    } else if string(cmd_args[0]) == "ea" {
	      if len(cmd_args) >=1 {
	        if err := client.ActionExec(cmd_args[1], []byte("testparam")); err != nil {
						fmt.Println(err)
						return
					}
	      }
	    } else if string(cmd_args[0]) == "es" {
	      if len(cmd_args) >=1 {
	        result, err := client.ServiceExec(cmd_args[1], []byte("testparam"))
					if err != nil {
						fmt.Println(err)
						return
					}
	        fmt.Println(string(result))
	      }
	    }
	  }
	}
}
