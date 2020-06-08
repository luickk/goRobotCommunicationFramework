/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_node

import (
  "os"
  "log"
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

var (
  InfoLogger    *log.Logger
  WarningLogger *log.Logger
  ErrorLogger   *log.Logger
)

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
  InfoLogger.Println("handleConnection started")
  defer conn.Close()

  for {
    byteData := make([]byte, tcpConnBuffer)
    n, err := bufio.NewReader(conn).Read(byteData)
    if err != nil {
      ErrorLogger.Println("handleConnection routine reading error")
      ErrorLogger.Println(err)
    }
    byteData = byteData[:n]
    // data := string(byteData)

    delimSplitByteData := bytes.Split(byteData, []byte("\r"))


    // iterating ovre conn read buffer array, split by backslash r
    for _, cmdByte := range delimSplitByteData {

      ptype, name, operation, payload := rcf_util.ParseNodeReadProtocol(cmdByte)

      if ptype != "" && name != "" {
        if ptype == "topic" {
          if operation == "publish" {
            InfoLogger.Println("handleConnection data published")
            TopicPublishData(node, name, payload)
          } else if operation == "pull" {
            InfoLogger.Println("handleConnection data pulled")
            nmsg,err := strconv.Atoi(string(payload))
            if err != nil {
              WarningLogger.Println("handleConnection data pull conversion error")
            } else {
              TopicPullData(node, conn, name, nmsg)
            }
          } else if operation == "subscribe" {
            InfoLogger.Println("handleConnection topic subsc")
            TopicAddListenerConn(node, name, conn)
          } else if operation == "create" {
            InfoLogger.Println("handleConnection topic created")
            TopicCreate(node, name)
          } else if operation == "list" {
            InfoLogger.Println("handleConnection topic listed")
            conn.Write(append([]byte(">info-list_topics-req-"),[]byte(strings.Join(NodeListTopics(node), ",")+"\r")...))
          }
        } else if ptype == "action" {
          if operation == "exec" {
            InfoLogger.Println("handleConnection action execed")
            ActionExec(node, name, payload)
          }
        } else if ptype == "service" {
          if operation == "exec" {
            InfoLogger.Println("handleConnection service execed")
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
  InfoLogger.Println("topicHandler called")
  for {
    select {
      case topicListener := <-node.topicListenerConnCh:
          InfoLogger.Println("topicHandler listener added")
          node.topicListenerConns = append(node.topicListenerConns, topicListener)
      case pullRequest := <-node.topicPullCh:
        InfoLogger.Println("topicHandler data pulled")
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
        InfoLogger.Println("topicHandler data pushed")

        if rcf_util.TopicsContainTopic(node.topics, topicMsg.topicName){

          node.topics[topicMsg.topicName] = append(node.topics[topicMsg.topicName], topicMsg.msg)

          // check if topic exceeds topic cap limits
          if len(node.topics[topicMsg.topicName]) > topicCapacity {
            topicOverhead := len(node.topics[topicMsg.topicName])-topicCapacity
            // slicing size of slice to right size‚
            node.topics[topicMsg.topicName] = node.topics[topicMsg.topicName][topicOverhead:]
          }

          // check if topic, which data is pushed to, has a listening conn
          for _,topicListener := range node.topicListenerConns {
            topicOnlyName, _ := rcf_util.SplitServiceToNameId(topicListener.topicName)
            if topicOnlyName == topicMsg.topicName {
              // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
              topicListener.listeningConn.Write(append(append([]byte(">topic-"+topicListener.topicName+"-sub-"),[]byte(topicMsg.msg)...), []byte("\r")...))
            }
          }
        }

        case topicCreateName := <- node.topicCreateCh:
          if rcf_util.TopicsContainTopic(node.topics, topicCreateName) {
          } else {
            node.topics[topicCreateName] = [][]byte{}
            InfoLogger.Println("topicHandler topic created")
          }
    }
    time.Sleep(time.Duration(nodeFreq))
  }
}

func actionHandler(nodeInstance Node) {
  InfoLogger.Println("actionHandler started")
  for {
    select {
    case action := <- nodeInstance.actionCreateCh:
      nodeInstance.actions[action.actionName] = action.actionFunction
      InfoLogger.Println("actionHandler action created")
    case actionExec := <- nodeInstance.actionExecCh:
      InfoLogger.Println("actionHandler action execed")
      if _, ok := nodeInstance.actions[actionExec.actionName]; ok {
        actionFunc := nodeInstance.actions[actionExec.actionName]
        go actionFunc(actionExec.params,nodeInstance)
      } else {
        InfoLogger.Println("actionHandler action execed not found")
      }
    }
    time.Sleep(time.Duration(nodeFreq))
  }
}

func serviceHandler(nodeInstance Node) {
  InfoLogger.Println("actionHandler called")
  for {
    select {
      case service := <- nodeInstance.serviceCreateCh:
        nodeInstance.services[service.serviceName] = service.serviceFunction
        InfoLogger.Println("serviceHandler service created")
      case serviceExec := <-nodeInstance.serviceExecCh:
        InfoLogger.Println("serviceHandler service execed(queued)")
        serviceOnlyName, _ := rcf_util.SplitServiceToNameId(serviceExec.serviceName)
        if _, ok := nodeInstance.services[serviceOnlyName]; ok {
          go func() {
            serviceResult := append(nodeInstance.services[serviceOnlyName](serviceExec.params, nodeInstance), "\r"...)
			      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
            serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), serviceResult...))
            InfoLogger.Println("serviceHandler service returned and wrote payload")
          }()
        } else {
		      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
          serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), []byte(serviceExec.serviceName+" not found \r")...))
          InfoLogger.Println("serviceHandler service not found")
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
  InfoLogger = log.New(os.Stdout, "[NODE] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
  WarningLogger = log.New(os.Stdout, "[NODE] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
  ErrorLogger = log.New(os.Stdout, "[NODE] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

  rcf_util.InfoLogger = InfoLogger
  rcf_util.WarningLogger = WarningLogger
  rcf_util.ErrorLogger = ErrorLogger

  go topicHandler(node)

  go actionHandler(node)

  go serviceHandler(node)
  InfoLogger.Println("Init handlers routine started")

  var port string = ":"+strconv.Itoa(node.id)

  l, err := net.Listen("tcp4", port)

  if err != nil {
    
  }

  defer l.Close()
  for {
    conn, err := l.Accept()
    if err != nil {
      ErrorLogger.Println("Init client conn accept error")
      ErrorLogger.Println(err)
    }
    go handleConnection(node, conn)
    InfoLogger.Println("Init new conn handler routine started")
  }
}

func NodeHalt() {
  for{time.Sleep(1*time.Second)}
  InfoLogger.Println("NodeHalt called")
}

func TopicAddListenerConn(node Node, topicName string, conn net.Conn) {
  topicListenerConn := new(topicListenerConn)
  topicListenerConn.topicName = topicName
  topicListenerConn.listeningConn = conn
  node.topicListenerConnCh <- *topicListenerConn
  topicListenerConn = nil
  InfoLogger.Println("TopicAddListenerConn called")
}

func NodeListTopics(node Node) []string{
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
  InfoLogger.Println("NodeListTopics called")
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
  InfoLogger.Println("TopicPullData called")
}

func TopicPublishData(node Node, topicName string, tdata []byte) {
  topicMsg := new(topicMsg)
  topicMsg.topicName = topicName
  topicMsg.msg = tdata
  node.topicPushCh <- *topicMsg
  topicMsg = nil
  InfoLogger.Println("TopicPublishData called")
}

// create command&control topic
func TopicCreate(node Node, topicName string) {
  topicName = rcf_util.ApplyNamingConv(topicName)

  node.topicCreateCh <- topicName
  InfoLogger.Println("TopicCreate called")
}

func ActionCreate(node Node, actionName string, actionFunc actionFn) {
    actionName = rcf_util.ApplyNamingConv(actionName)
    newAction := new(action)
    newAction.actionName = actionName
    newAction.actionFunction = actionFunc
    node.actionCreateCh <- *newAction
    newAction = nil
    InfoLogger.Println("ActionCreate called")
}

func ActionExec(node Node, actionName string, actionParams []byte) {
  actionExec := new(actionExec)
  actionExec.actionName = actionName
  actionExec.params = actionParams
  node.actionExecCh <- *actionExec
  actionExec = nil
  InfoLogger.Println("ActionExec called")
}

func ServiceCreate(node Node, serviceName string, serviceFunc serviceFn) {
    serviceName = rcf_util.ApplyNamingConv(serviceName)
    service := new(service)
    service.serviceName = serviceName
    service.serviceFunction = serviceFunc
    node.serviceCreateCh <- *service
    service = nil
    InfoLogger.Println("ServiceCreate called")
}

func ServiceExec(node Node, conn net.Conn, serviceName string, serviceParams []byte) {
  serviceExec := new(serviceExec)
  serviceExec.serviceName = serviceName
  serviceExec.serviceCallConn = conn
  serviceExec.params = serviceParams
  node.serviceExecCh <- *serviceExec
  serviceExec = nil
  InfoLogger.Println("ServiceExec called")
}