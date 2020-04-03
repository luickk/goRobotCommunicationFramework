# Robot Communication Framework

The RCF is a framework for data distribution, which is the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages and the C/C++ complexity overhead while still maintaining speed and **safe** thread/ lang. standards, thanks to the [go](https://golang.org/) lang.

Examples can be foune in ./examples.

# Concept

The primary communication interface is a node, in contrast to ROS, or various other robot platforms, the node is only a object instance and does not contain any code. A node resembles the platform for services, actions and topics.

Every node has a port number which also resembles the node ID, but no actual name, so every node is represented through a number. This is a major part of the concept to reduce complexity since there is no need for internal node communication to resolve addresses.

# Installation

Installation via. command line: <br>

`go get github.com/cy8berpunk/robot-communication-framework` <br>

installation from code

`
import (
    "fmt"
    "github.com/cy8berpunk/robot-communication-framework"
)
`
# Go-Routine Memory synchronization

Since sharing memory is a very complicated and difficult thing to get right without overcomplicating things, Go channels are used, according to Golang's motto "Share memory by communicating, don't communicate by sharing memory"!

## Topics

Every topic represents a communication channel, from which data can be pulled from or pushed onto or to which can be listened.
The topic communication is split up into msg's, every msg represents a byte array pushed to the topic. Every node has a topic msg capacity, so only the last x msg's are stored. There are no variable assignments, a msg can represent a single value or anything else encoded into a byte array. If a topic msg structure is needed, the `glob` methods serialize a string map and use the serialized maps as msgs and as such enable a structured and more generic way to use topics.
A topic is meant to share command & control or sensor data, hence data that needs to be accurate and which does not require high bandwith, since a topic rely's on tcp sockets to communicate.
A topic can be identified via its name and the node(node ID) which it is hosted by.

## Actions

An action is a function that can be executed, with parameters, by nodes or node clients. Since they are not meant to do calculations but to provide node side functionality for the clients, they can only be called without a return value.
Has to be declared on the node side.

## Services

A service is a function that can be executed with parameters and process for a certain amount of time, to finally return a result in form of a byte array.
Has to be declared on the node side.

## Protocols

#### Delimiter

Single commands are separated by a "\n". Msgs, encoded in commands(protocol based) are also separated by a "\n" (for example the glob maps).

#### Node

`><type>-<name>-<operation>-<paypload byte slice>`

#### Client Read Protocol

`><type>-<name>-<len(msgs)>-<paypload(msgs) byte slice>`

## Functions

### Client

`Node_open_conn(node_id int) net.Conn` // opens tcp connection to node and returns it <br> 
`Node_close_conn(conn net.Conn)`// closes conn <br> 

#### Topics
------

##### Publish
`Topic_publish_raw_data(conn net.Conn, topic_name string, data []byte)` <br>
Publishes byte slice type msg to topic with topic name <br>
`Topic_publish_string_data(conn net.Conn, topic_name string, data string)`<br>
Publishes string type msg to topic with topic name <br>
`Topic_publish_glob_data(conn net.Conn, topic_name string, data data map[string]string)`<br>
Publishes serialized map type msg to topic with topic name <br>

##### Pull

`Topic_pull_raw_data(conn net.Conn, nmsgs int, topic_name string) [][]byte`<br>
Pulls nmsgs amount of, byte slice type msgs from topic name  <br>
`Topic_pull_string_data(conn net.Conn, nmsgs int, topic_name string) []string`<br>
Pulls nmsgs amount of, string type msgs from topic name <br>
`Topic_pull_glob_data(conn net.Conn, nmsgs int, topic_name string) map[string]string`<br>
Pulls nmsgs amount of, serialized map type msgs from topic name  <br> 

##### Subscribe
`Topic_raw_raw_subscribe(conn net.Conn, topic_name string) <-chan []byte`<br>
Listens to every new msg(byte slice type) pushed to topic with topic name and pushes it to returned channel <br>
`Topic_raw_string_subscribe(conn net.Conn, topic_name string) <-chan string`<br>
Listens to every new msg(string type) pushed to topic with topic name and pushes it to returned channel <br>
`Topic_raw_glob_subscribe(conn net.Conn, topic_name string) <-chan map[string]string`<br>
Listens to every new msg(serialized map type) pushed to topic with topic name and pushes it to returned channel <br>

##### Util
`Topic_create(conn net.Conn, topic_name string)`
Creates new topic on node with given node conn <br>
`Topic_list(conn net.Conn) []string`
Lists provided topics of given node conn <br>

#### Actions/ Services (have to be provided by Node)
`Action_exec(conn net.Conn, action_name string, params []byte)`<br>
Executes action provided by node of given node conn <br>
`Service_exec(conn net.Conn, service_name string, params []byte) []byte`<br>
Executes service provided by node of given node conn. returned byte slice equals service function result`<br>

### Node 

`Create(node_id int) Node`<br>
Creates node object and initiates node struct<br>
`Init(node Node)`<br>
Calls service,action,topic and connection handlers<br>

#### Services

`Service_create(node Node, service_name string, service_func service_fn)`<br>
Creates service on node. the service is then provided by the node and can be executed by clients <br>
`Service_exec(node Node, conn net.Conn, service_name string, service_params []byte)`<br>
Executes service, equals execution by client  <br>

#### Actions

`Action_create(node Node, action_name string, action_func action_fn)`<br>
Creates action on node. the action is then provided by the node and can be executed by clients <br>
`Action_exec(node Node, action_name string, action_params []byte)`<br>
Executes action, equals execution by client   <br>

#### Topics

`Topic_create(node Node, topic_name string) ` <br>
Creates topic with topic name on node <br>
`Topic_publish_data(node Node, topic_name string, tdata []byte)` <br>
Publishes byte slice type msg to topic with topic name <br> 
`Topic_pull_data(node Node, topic_name string, nmsgs int) [][]byte` <br>
Pulls nmsgs amount of, byte slice type msgs from topic name  <br>

`Node_list_topics(node Node) []string` <br>
Lists all topics provided by node <br>
`Topic_add_listener_conn(node Node, topic_name string, conn net.Conn) ` <br> 
Adds given connection to subscribed connection list, for topic name. <br>
