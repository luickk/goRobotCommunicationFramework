package main

import (
	"fmt"
  "os"
  "strconv"
  "math/rand"

	"goRobotCommunicationFramework/rcfNodeClient"
)

func main() {
  topicNameArg := os.Args[1]
	// opening connection(tcp client) to node with id(port) 30
	client, err := rcfNodeClient.New(47)
  if err != nil {
    fmt.Println(err)
  }
  // initiating topic listener
	// returns channel which every new incoming element/ msg is pushed to
	topicListener, err := client.TopicDataSubscribe(topicNameArg)
	if err != nil {
		fmt.Println(err)
		return
	}

  // smaple loop
  for {
    // select statement to wait for new incoming elements/msgs from listened to topic
    select {
    // if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
    case data := <-topicListener:
    	// converting altitude element/ msg which is encoded as string to integer
    	// removing spaces before
    	// printing new altitude, pushed to topic
    	fmt.Println("- floodClientSubscribe(" + topicNameArg + "): " + string(data))
      dataVal, err := strconv.Atoi(string(data))
      if err != nil {
        fmt.Println(err)
        break
      }

    	// checking if new altitude is equal tp 99 for example purposes
    	if dataVal == 99 {
    		rand := strconv.Itoa(rand.Intn(1000000))
    		go func() {
    			res, err := client.ServiceExec("testDelayService", []byte("randTestParamFromSub"+rand))
					if err != nil {
						fmt.Println(err)
						return
					}
    			fmt.Println("results delay: " + string(res))

    			client.ActionExec("testAction", []byte("randTestParamFromSub"+rand))
    		}()
    	}
    }
  }

}
