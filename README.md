# Robot Communication Framework

The Robot Communication Framework is a framework for data distribution and computing which is the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages and the C/C++ complexity overhead while still maintaining speed and **safe** thread management standards thanks to the [go](https://golang.org/) lang.

## Introduction

The primary communication interface resembles the node instance, in contrast to ROS or various other robot platforms. A node resembles the platform for services, actions and topics which can be accessed by the rcf clients.
Every node has a port number which also resembles the node Id but no actual name so every node is represented by a number. This is a major part of the concept to reduce complexity since there is no need for internal node communication to resolve addresses.

# Philosophy

Communication and concurrency is one of the most prominent challenges when developing a robotic platform. Since every robot can have multiple kinds of sensors and control interfaces it has to interact and to communicate with, communication is an absolute indispensable part of the foundation of every robot.
Considering such rather complex requirements, the probability of failure increases due to the increase in complexity. So the goal of this project is not just to keep the complexity low but to increase the reliability and maintainability of the framework. This is achieved by using only core feature and shrinking the dependencies to an absolute minimum. Another aspect that needs to be considered is the chosen programming language and its reliability/ dependency. This projects uses go, with absolutely no dependencies, using only core features for concurrency and networking. Another great thing about go is that it does not require a virtual runtime of any kind and integrates more or less bare metal in the system(as for example C/C++ does). A further great feature is its concurrency support which is achieved by eliminating parallelization and replacing it with a channel interface with which data can be shared between functions. That is done by guaranteeing that the data is handled, if not done so, the program will deadlock to prohibit any memory parallelization issues. According to Golang's motto "Share memory by communicating, don't communicate by sharing memory"!. 

# Installation

Installation via. command line: <br>

1. `cd Users/<username>/go/src/ ` <br>

2. `git clone https://github.com/cy8berpunk/rcf` <br>

# Tutorial

## Simple test setup

First off the node has to be started:

1. `cd examples/ ` <br> 
2. `go run node.go` <br>
 
Next the publisher is launched, which also creates the topic and then begins to publish random sample data with a constant frequency:

3. `go run publisher_string.go` <br>

The last step is to launch the worker which subscribes to the topic created by the publisher and prints the data:

4. `go run worker_string.go` <br>

## Full endurance test

Test all functions of the library with an high node refresh frequency:

1. `cd tools/ ` <br> 

2. `bash enduranceTest.sh` <br>

# API reference

## Specifications

Every topic represents a communication channel from which data can be pulled from or pushed onto or to which can be listened for new msgs(subscribed).
The topic communication is split up into msg's, every msg represents a byte array pushed to the topic. Every node has a topic msg capacity, so only the last x msg's are stored. There are no variable assignments, a msg can represent a single value or anything else encoded into a byte array. If a topic msg structure is needed the `glob` methods serialize a string map and use the serialized maps as msgs and as such enable a structured and more generic way to use topics.
A topic is meant to share command & control or sensor data, hence data that needs to be accurate and which does not require high bandwith since a topic relies on tcp sockets to communicate.
A topic can be identified via. its name and the node(node ID) which it is hosted by.

## Actions

An action is a function that can be execute with parameters by nodes or node clients. Since they are not meant to do calculations but to provide node side functionality to the clients they can only be called without a return value.
Has to be initiated on the node.

## Services

A service is a function that can be executed with parameters and asynchronously, respectively processes for a certain amount of time while still returning the result(payload) to the service call.
Has to be initiated on the node.

## Protocols

#### Delimiter

A single instrucion, respectively every single tcp blocks is separated by a "\r"

#### Protocol(Instruction)

`><type>-<name>-<operation>-<lengths(amount of which resembles amount of payload elements delim:",". use is optional)>-<paypload byte slice>`

### Package Functions

#### Node

- `Create(nodeId int) Node` <br>
Creates the node struct object and declares all struct elements
- `Init(node Node)` <br>
Initiates all node struct elements
- `NodeHalt()` <br>
Pauses the node 
- `NodeListTopics(node Node) []string` <br>
Lists all created topics
- `TopicPublishData(node Node, topicName string, tdata []byte)` <br>
Publishes given raw msg to given node 
- `TopicCreate(node Node, topicName string)` <br>
Creates a topic with given topic name
- `ActionCreate(node Node, actionName string, actionFunc actionFn)` <br>
Creates action on given node
```
  rcfNode.ActionCreate(nodeInstance, "testAction", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED.")
    println(string(params))
  })
```
- `ServiceCreate(node Node, serviceName string, serviceFunc serviceFn)` <br>
Creates service on given node. <br>
Example:
```
  rcfNode.ServiceCreate(nodeInstance, "testServiceDelay", func(params []byte, n rcfNode.Node) []byte {
    NserviceExeced += 1
    fmt.Println("---- Service delay TEST EXECUTED("+strconv.Itoa(NserviceExeced)+" times). Param: "+string(params))
    time.Sleep(1*time.Second)
    return params
  })
``` 
- `ServiceExec(node Node, conn net.Conn, serviceName string, serviceParams []byte)` <br>
Executes service on given node with given name and given parameters


#### Node Client

- `TopicPullRawData(clientStruct client, topicName string, nmsgs int) [][]byte` <br>
Pulls n amount of raw msgs from given topic and returns them. <br>
- `TopicRawDataSubscribe(clientStruct client, topicName string) chan []byte` <br>
Subscribes to given topic and returns live string msgs by pushing them into the returned channel. A subscription to a topic resembles a live data stream of new msgs pushed to the topic. <br>
- `TopicPublishRawData(clientStruct client, topicName string, data []byte)`<br>
Pushes given raw msg to the given topic. <br>
- `TopicPublishStringData(clientStruct client, topicName string, data string)` <br>
Pushes given string msg to the given topic. <br>
- `TopicPullStringData(clientStruct client, nmsgs int, topicName string) []string` <br>
Pulls n amount of string msgs from given topic and returns them. <br>
- `TopicStringDataSubscribe(clientStruct client, topicName string) <-chan string` <br>
Subscribes to given topic and returns live string msgs by pushing them into the returned channel. <br>
- `TopicPublishGlobData(clientStruct client, topicName string, data map[string]string) ` <br>
Publishes given glob map (protocol conform: `map[string]string`) on the given node.<br>
- `TopicPullGlobData(clientStruct client, nmsgs int, topicName string) []map[string]string ` <br>
Pulls n amount of serialized glob map msgs from given topic and returns them as maps.<br>
- `TopicGlobDataSubscribe(clientStruct client, topicName string) <-chan map[string]string` <br>
Subscribes to given topic and returns live glob maps by pushing them into the returned channel. <br>
- `ActionExec(clientStruct client, actionName string, params []byte` <br>
Executes given action on the connected node<br>
- `TopicCreate(clientStruct client, topicName string)` <br>
Creates topic with given name on connected node.<br>
- `TopicList(clientStruct client, connChannel chan []byte) []string`<br>
Sends topic list request to node and returns all on the node created channels.<br>
