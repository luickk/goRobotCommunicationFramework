/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_node

import (
	"fmt"
	"bufio"
	"bytes"
	"net"
	"time"
	"strings"
	"strconv"
	"rcf/rcf-util"
)

// node msg/ element history length
var topicCapacity = 5

// frequency with which nodes handlers are regfreshed
var nodeFreq = 0

var tcpConnBuffer = 1024

type topicMsg struct {
  topicName string
  msg []byte
}

type topicPullReq struct {
  conn net.Conn
  topicName string
  nmsg int
}

type topicListenerConn struct {
  listeningConn net.Conn
  topicName string
}

type action struct {
  actionName string
  actionFunction actionFn
}

type service struct {
  serviceName string
  serviceFunction serviceFn
}

type serviceExec struct {
  serviceName string
  serviceCallConn net.Conn
  params []byte
}

type actionExec struct {
  actionName string
  params []byte
}
// node struct
type Node struct {
  // id or port of node
  id int

  // key: topic name, value: stack slice
  topics map[string][][]byte

  topicPushCh chan topicMsg

  topicCreateCh chan string

  topicListenerConnCh chan topicListenerConn

  topicPullCh chan topicPullReq

  topicListenerConns []topicListenerConn

  // action map with first key(action name) value(anon type action func) pair
  actions map[string]actionFn

  actionCreateCh chan action

  actionExecCh chan actionExec


  // service map with first key(service name) value(anon type services func) pair
  services map[string]serviceFn

  serviceCreateCh chan service

  serviceExecCh chan serviceExec
}

// general action function type
type actionFn func (params []byte, nodeInstance Node)

// general service function type
type serviceFn func (params []byte, nodeInstance Node) []byte


// handles every incoming node client connection
func handleConnection(node Node, conn net.Conn) {
  defer conn.Close()

  for {
    byteData := make([]byte, tcpConnBuffer)
    n, err_handle := bufio.NewReader(conn).Read(byteData)
    byteData = byteData[:n]
    // data := string(byteData)

    delimSplitByteData := bytes.Split(byteData, []byte("\r"))
    // delim_split_data := strings.Split(data, "\r")


    // iterating ovre conn read buffer array, split by backslash r
    for _, cmd_b := range delimSplitByteData {
      // cmd := delim_split_data[i]

      if err_handle != nil {
        fmt.Println("/[node] ", err_handle)
        return
      }

      // Node read protocol:
      // ><type>-<name>-<operation>-<paypload byte slice>
      ptype, name, operation, payload := rcf_util.ParseNodeReadProtocol(cmd_b)

      if ptype != "" && name != "" {
        // fmt.Println(cmd)
        // fmt.Println(ptype+","+name+","+operation+","+string(payload) + ". end")
        if ptype == "topic" {
          if operation == "publish" {
            TopicPublishData(node, name, payload)
          } else if operation == "pull" {
            nmsg,_ := strconv.Atoi(string(payload))
            TopicPullData(node, conn, name, nmsg)
          } else if operation == "subscribe" {
            TopicAddListenerConn(node, name, conn)
          } else if operation == "create" {
            TopicCreate(node, name)
          } else if operation == "list" {
            conn.Write(append([]byte(">info-list_topics-req-"),[]byte(strings.Join(NodeListTopics(node), ",")+"\r")...))
          }
        } else if ptype == "action" {
          if operation == "exec" {
            ActionExec(node, name, payload)
          }
        } else if ptype == "service" {
          if operation == "exec" {
            ServiceExec(node, conn, name, payload)
          }
        }
      }
    }
    // data = ""
    byteData = []byte{}
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topicHandler(node Node) {
  for {
    select {
      case topicListener := <-node.topicListenerConnCh:
          node.topicListenerConns = append(node.topicListenerConns, topicListener)
      case pullRequest := <-node.topicPullCh:
        var byteData [][]byte 
        topicOnlyName, _ := rcf_util.SplitServiceToNameId(pullRequest.topicName)
        if pullRequest.nmsg >= len(node.topics[topicOnlyName]){
          byteData = node.topics[topicOnlyName]
        } else {
          byteData = node.topics[topicOnlyName][:pullRequest.nmsg]
        }

        if(pullRequest.nmsg<=1) {
          // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
          if len(byteData) >= 1 {
            pullRequest.conn.Write(append(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), byteData[0]...), []byte("\r")...))
          } else {
            pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...))
          }
        } else {
          if len(byteData) >= 1 {
            tdata := append(bytes.Join(byteData, []byte("\nm")), []byte("\r")...)
            // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
            pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), tdata...))
          } else {
            pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...))
          }
        }
      case topicMsg := <-node.topicPushCh:

        if rcf_util.TopicsContainTopic(node.topics, topicMsg.topicName){

          node.topics[topicMsg.topicName] = append(node.topics[topicMsg.topicName], topicMsg.msg)

          // check if topic exceeds topic cap limits
          if len(node.topics[topicMsg.topicName]) > topicCapacity {
            topic_overhead := len(node.topics[topicMsg.topicName])-topicCapacity
            // slicing size of slice to right size‚
            node.topics[topicMsg.topicName] = node.topics[topicMsg.topicName][topic_overhead:]
          }

          // check if topic, which data is pushed to, has a listening conn
          for _,topicListener := range node.topicListenerConns {
            if topicListener.topicName == topicMsg.topicName {
              // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
              topicListener.listeningConn.Write(append(append([]byte(">topic-"+topicMsg.topicName+"-sub-"),[]byte(topicMsg.msg)...), []byte("\r")...))
            }
          }
        }

        case topicCreateName := <- node.topicCreateCh:
          fmt.Println("+[topic] ", topicCreateName)
          if rcf_util.TopicsContainTopic(node.topics, topicCreateName) {
            fmt.Println("/[topic] ", topicCreateName)
          } else {
            node.topics[topicCreateName] = [][]byte{}
          }
    }
    time.Sleep(time.Duration(nodeFreq))
  }
}

func actionHandler(nodeInstance Node) {
  for {
    select {
    case action := <- nodeInstance.actionCreateCh:
      nodeInstance.actions[action.actionName] = action.actionFunction
    case actionExec := <- nodeInstance.actionExecCh:
      if _, ok := nodeInstance.actions[actionExec.actionName]; ok {
        action_func := nodeInstance.actions[actionExec.actionName]
        go action_func(actionExec.params,nodeInstance)
      } else {
        fmt.Println("/[action] ", actionExec)
      }
    }
    time.Sleep(time.Duration(nodeFreq))
  }
}

func serviceHandler(nodeInstance Node) {
  for {
    select {
      case service := <- nodeInstance.serviceCreateCh:
        nodeInstance.services[service.serviceName] = service.serviceFunction
      case serviceExec := <-nodeInstance.serviceExecCh:
        serviceOnlyName, _ := rcf_util.SplitServiceToNameId(serviceExec.serviceName)
        if _, ok := nodeInstance.services[serviceOnlyName]; ok {
          go func() {
            service_result := append(nodeInstance.services[serviceOnlyName](serviceExec.params, nodeInstance), "\r"...)

			      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
            serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), service_result...))
          }()
        } else {
          fmt.Println("/[service] ", serviceOnlyName)
		      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
          serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), []byte(serviceExec.serviceName+" not found \r")...))
        }
      time.Sleep(time.Duration(nodeFreq))
    }
  }
}



// creating node instance struct
func Create(nodeId int) Node{
  // key: topic name, value: stack slice
  topics := make(map[string][][]byte)

  topicPushCh := make(chan topicMsg)

  topicCreateCh := make(chan string)

  topicListenerConnCh := make(chan topicListenerConn)

  topicPullCh := make(chan topicPullReq)

  topicListenerConns := make([]topicListenerConn,0)

  // action map with first key(action name) value(anon action func) pair
  actions := make(map[string]actionFn)

  actionCreateCh := make(chan action)

  actionExecCh := make(chan actionExec)

  services := make(map[string]serviceFn)

  serviceCreateCh := make(chan service)

  serviceExecCh := make(chan serviceExec)

  return Node{nodeId, topics, topicPushCh, topicCreateCh, topicListenerConnCh, topicPullCh, topicListenerConns, actions, actionCreateCh, actionExecCh, services, serviceCreateCh, serviceExecCh}
}

// createiating node with given id
// returns createiated node instance to enable direct service and topic operations
func Init(node Node) {
  fmt.Println("+[node] ", node.id)

  go topicHandler(node)

  go actionHandler(node)

  go serviceHandler(node)

  var port string = ":"+strconv.Itoa(node.id)

  l, err := net.Listen("tcp4", port)

  if err != nil {
    fmt.Println("/[node] ", err)
  }

  defer l.Close()
  for {
    conn, err_handle := l.Accept()
    if err_handle != nil {
      fmt.Println("/[node] ",err_handle)
    }
    go handleConnection(node, conn)
  }
}

func NodeHalt() {
  for{time.Sleep(1*time.Second)}
}

func TopicAddListenerConn(node Node, topicName string, conn net.Conn) {
  topicName = rcf_util.ApplyNamingConv(topicName)
  fmt.Println("-> sub ", topicName)
  topicListenerConn := new(topicListenerConn)
  topicListenerConn.topicName = topicName
  topicListenerConn.listeningConn = conn
  node.topicListenerConnCh <- *topicListenerConn
  topicListenerConn = nil
}

func NodeListTopics(node Node) []string{
  fmt.Println("listing topics")
  keys := make([]string, 0, len(node.topics))
  for k, v := range node.topics {
    v=v
    keys = append(keys, k)
  }
  if len(node.topics) > 0 {
    return keys
  } else if len(node.topics) == 0 {
    return []string{"none"}
  }
  return []string{"none"}
}

// pushes pull request to pull topic channel
// request is handled in the topic handler
func TopicPullData(node Node, conn net.Conn, topicName string, nmsg int) {
  topicPullReq := new(topicPullReq)
  topicPullReq.topicName = topicName
  topicPullReq.nmsg = nmsg
  topicPullReq.conn = conn
  node.topicPullCh <- *topicPullReq
}

func TopicPublishData(node Node, topicName string, tdata []byte) {
  topicMsg := new(topicMsg)
  topicMsg.topicName = topicName
  topicMsg.msg = tdata
  node.topicPushCh <- *topicMsg
  fmt.Println("->[topic] ", topicName)
  topicMsg = nil
}

// create command&control topic
func TopicCreate(node Node, topicName string) {
  topicName = rcf_util.ApplyNamingConv(topicName)

  node.topicCreateCh <- topicName
}

func ActionCreate(node Node, actionName string, action_func actionFn) {
    actionName = rcf_util.ApplyNamingConv(actionName)
    newAction := new(action)
    newAction.actionName = actionName
    newAction.actionFunction = action_func
    node.actionCreateCh <- *newAction
    newAction = nil
}

func ActionExec(node Node, actionName string, action_params []byte) {
  actionExec := new(actionExec)
  actionExec.actionName = actionName
  actionExec.params = action_params
  node.actionExecCh <- *actionExec
  actionExec = nil
}

func ServiceCreate(node Node, serviceName string, service_func serviceFn) {
    serviceName = serviceName
    service := new(service)
    service.serviceName = serviceName
    service.serviceFunction = service_func
    node.serviceCreateCh <- *service
    service = nil
}

func ServiceExec(node Node, conn net.Conn, serviceName string, service_params []byte) {
  serviceExec := new(serviceExec)
  serviceExec.serviceName = serviceName
  serviceExec.serviceCallConn = conn
  serviceExec.params = service_params
  node.serviceExecCh <- *serviceExec
  serviceExec = nil
}