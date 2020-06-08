package rcf_node_client

import(
	"net"
	"strconv"
	"strings"
	"bufio"
  "bytes"
  "log"
  "rcf/rcf-util"
  "os"
)

type dataRequest struct {
  // name always includes optional id
  Name string
  Op string
  Fulfilled bool

  ReturnedPayload chan []byte
  PullOpReturnedPayload chan [][]byte
}

type client struct {
  Conn net.Conn
  TopicContextRequests chan dataRequest 
  ServiceContextRequests chan dataRequest 
}

var tcpConnBuffer = 1024
var topicContextMsgs chan []byte
var serviceContextMsgs chan []byte

var (
  InfoLogger    *log.Logger
  WarningLogger *log.Logger
  ErrorLogger   *log.Logger
)

// parses information and pushes them to byte channel 
func connHandler(conn net.Conn, topicContextMsgs chan []byte, serviceContextMsgs chan []byte) {
  InfoLogger.Println("connHandler started")
  for {
    data := make([]byte, tcpConnBuffer)
    n, err := bufio.NewReader(conn).Read(data)
    if err != nil {
      ErrorLogger.Fatalln("connHandler socket read err")
      ErrorLogger.Fatalln(err)
    }
    data = data[:n]
    splitRData := bytes.Split(data, []byte("\r"))
    for _, data := range splitRData {
      if len(data)>=1 {
        ptype, _, _, _ := rcf_util.ParseNodeReadProtocol(data)
        if ptype != "" {
          if ptype == "topic" {
              InfoLogger.Println("connHandler topic msg parsed")
              topicContextMsgs <- data
            } else if ptype == "service" {
              InfoLogger.Println("connHandler service msg parsed")
              serviceContextMsgs <- data
          }
        }
      }
    }
  }
}

func topicHandler(conn net.Conn, topicContextMsgs chan []byte, topicRequests chan dataRequest) {
  requests := []dataRequest{} 
  InfoLogger.Println("topicHandler started")
  for {
    select {
      case data := <-topicContextMsgs:
        //only for parsing purposes
        dataString := string(data)
        if strings.SplitN(dataString, "-", 4)[2] == "pull"{
          for i, request := range requests {
            if request.Op == "pull" && request.Fulfilled == false {
              InfoLogger.Println("topicHandler pull handled")
              payloadMsgs := ParseTopicPulledRawData(data, request.Name)
              request.PullOpReturnedPayload <- payloadMsgs  
              requests[i].Fulfilled = true
            }
          }
        } else if strings.SplitN(dataString, "-", 4)[2] == "sub" {
          for _, request := range requests {
            if request.Op == "sub" && request.Fulfilled == false{
              var payload []byte
              if strings.Split(dataString, "-")[1] == request.Name {
                InfoLogger.Println("topicHandler subscribed handled")
                payload = bytes.SplitN(data, []byte("-"), 4)[3]
                request.ReturnedPayload <- payload
                // requests[i].Fulfilled = true
              }
            }
          }
        }
      case request := <-topicRequests:
        InfoLogger.Println("topicHandler request added")
        requests = append(requests, request)
    }
  }
}

func serviceHandler(conn net.Conn, serviceContextMsgs <-chan []byte, serviceRequests <-chan dataRequest) {
  requests := []dataRequest{}
  InfoLogger.Println("serviceHandler started")
  for {
    select {
      case data := <-serviceContextMsgs:
        if len(data) >= 0 {
          for i, request := range requests {
            if request.Fulfilled == false {
              InfoLogger.Println("serviceHandler service done executing")
              payload := ParseServiceReplyPayload(data, request.Name)
              if len(payload) != 0 { 
                InfoLogger.Println("serviceHandler service payload returned")
                request.ReturnedPayload <- payload
                requests[i].Fulfilled = true
              }
            }
          }
        }
      case request := <-serviceRequests:
        InfoLogger.Println("serviceHandler request added")
        requests = append(requests, request)
    }
  }
}

// executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func ServiceExec(clientStruct client, serviceName string, params []byte) []byte {
  InfoLogger.Println("ServiceExec service exec called")
  serviceId := rcf_util.GenRandomIntId()
  name := serviceName+","+strconv.Itoa(serviceId)

  request := new(dataRequest)
  request.Name = name
  request.Op = "exec"
  request.Fulfilled = false
  request.ReturnedPayload = make(chan []byte)
  clientStruct.ServiceContextRequests <- *request
  clientStruct.Conn.Write(append(append([]byte(">service-"+name+"-exec-"), params...), "\r"...))
  InfoLogger.Println("ServiceExec request sent")
  
  reply := false
  payload := []byte{}

  for !reply {
    select {
      case liveDataRes := <-request.ReturnedPayload:
        payload = liveDataRes
        InfoLogger.Println("ServiceExec Payload returned")
        reply = true      
        break
    }
  }
  return payload
}

func TopicPullRawData(clientStruct client, topicName string, nmsgs int) [][]byte {
  InfoLogger.Println("TopicPullRawData called")
  pullReqId := rcf_util.GenRandomIntId()
  name := topicName+","+strconv.Itoa(pullReqId)
  instructionSlice := append([]byte(">topic-"+name+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  clientStruct.Conn.Write(instructionSlice)


  request := new(dataRequest)
  request.Name = name
  request.Op = "pull"
  request.Fulfilled = false
  request.PullOpReturnedPayload = make(chan [][]byte)
  clientStruct.TopicContextRequests <- *request
  InfoLogger.Println("TopicPullRawData request sent")

  reply := false
  payload := [][]byte{}

  for !reply {
    select {
      case liveDataRes := <-request.PullOpReturnedPayload:
        payload = liveDataRes
        InfoLogger.Println("TopicPullRawData payload returned")
        reply = true      
        break
    }
  }
  return payload
}

func TopicRawDataSubscribe(clientStruct client, topicName string) chan []byte {
  InfoLogger.Println("TopicRawDataSubscribe called")
  pullReqId := rcf_util.GenRandomIntId()
  name := topicName+","+strconv.Itoa(pullReqId)
  clientStruct.Conn.Write([]byte(">topic-"+name+"-subscribe-\r"))

  request := new(dataRequest)
  request.Name = name
  request.Op = "sub"
  request.Fulfilled = false
  request.ReturnedPayload = make(chan []byte)
  clientStruct.TopicContextRequests <- *request
  InfoLogger.Println("TopicRawDataSubscribe request sent and channel returned")

  return request.ReturnedPayload
}


// function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed 
func connectToTcpServer(port int) (net.Conn, chan []byte, chan []byte) {
  InfoLogger.Println("connectToTcpServer called")
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))
  topicContextMsgs = make(chan []byte)
  serviceContextMsgs = make(chan []byte)

  go connHandler(conn, topicContextMsgs, serviceContextMsgs)

  if err != nil {
    ErrorLogger.Fatalln("connectToTcpServer could not connect to tcp server (node instance)")
    ErrorLogger.Fatalln(err)
  }
  // don't forget to close connection
  return conn, topicContextMsgs, serviceContextMsgs
}

// returns connection to node
func NodeOpenConn(nodeId int) client {
  InfoLogger = log.New(os.Stdout, "[CLIENT] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
  WarningLogger = log.New(os.Stdout, "[CLIENT] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
  ErrorLogger = log.New(os.Stdout, "[CLIENT] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

  rcf_util.InfoLogger = InfoLogger
  rcf_util.WarningLogger = WarningLogger
  rcf_util.ErrorLogger = ErrorLogger

  InfoLogger.Println("NodeOpenConn called")
  conn, topicContextMsgs, serviceContextMsgs := connectToTcpServer(nodeId)
  topicContextRequests := make(chan dataRequest)
  serviceContextRequests := make(chan dataRequest)

  
  go topicHandler(conn, topicContextMsgs, topicContextRequests)
  go serviceHandler(conn, serviceContextMsgs, serviceContextRequests)
  InfoLogger.Println("Init handler routines started")

  client := new(client)
  client.Conn = conn
  client.TopicContextRequests = topicContextRequests
  client.ServiceContextRequests = serviceContextRequests

  return *client
}

func NodeCloseConn(clientStruct client) {
  InfoLogger.Println("NodeCloseConn closed")
  clientStruct.Conn.Write([]byte("end\r"))
  clientStruct.Conn.Close()
}

func ParseServiceReplyPayload(data []byte, name string) []byte {  
  InfoLogger.Println("ParseServiceReplyPayload called")
  var payload []byte
  dataString := string(data)
  splitData := strings.Split(dataString, "-")
  
  if len(splitData) >= 2 {
    msgServiceName := splitData[1]
    if msgServiceName == name {
      InfoLogger.Println("ParseServiceReplyPayload payload returned")
      payload = bytes.SplitN(data, []byte("-"), 4)[3]
    }
  }
  return payload
}

// pulls x msgs from topic topic stack
func ParseTopicPulledRawData(data []byte, name string) [][]byte {
  InfoLogger.Println("ParseTopicPulledRawData called")
  var msgs [][]byte
  var payload []byte
  dataString := string(data)
  msgTopicName := strings.SplitN(dataString, "-", 4)[1]
  if msgTopicName == name {
    payload = bytes.SplitN(data, []byte("-"), 4)[3]
    splitPayload := bytes.Split(payload, []byte("\nm"))

    for _, splitPayloadMsg := range splitPayload {
      if len(splitPayloadMsg) >= 1 {
        InfoLogger.Println("ParseTopicPulledRawData payload returned")
        msgs = append(msgs, splitPayloadMsg)
      }
    }
  }
  return msgs
}

// pushes raw byte slice msg to topic msg stack
func TopicPublishRawData(clientStruct client, topicName string, data []byte) {
  sendSlice := append(append([]byte(">topic-"+topicName+"-publish-"),data...),"\r"...)
  clientStruct.Conn.Write(sendSlice)
  InfoLogger.Println("TopicPublishRawData called")
}


// pushes string msg to topic msg stack
func TopicPublishStringData(clientStruct client, topicName string, data string) {
  TopicPublishRawData(clientStruct, topicName, []byte(data))
}

// pulls x msgs from topic topic stack
func TopicPullStringData(clientStruct client, nmsgs int, topicName string) []string {
  InfoLogger.Println("TopicPullStringData called")
  var stringPayload []string
  payloadMsgs := TopicPullRawData(clientStruct, topicName, nmsgs)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      stringPayload = append(stringPayload, string(payloadMsg))
      InfoLogger.Println("TopicPullStringData string converted")
    }
  }

  return stringPayload
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicStringDataSubscribe(clientStruct client, topicName string) <-chan string {
  InfoLogger.Println("TopicStringDataSubscribe called")
  rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
  stringReturnListener := make(chan string)
  go func(stringReturnListener chan<- string ){
    for {
      select {
        case rawData := <-rawReturnListener:
          stringReturnListener <- string(rawData)
          InfoLogger.Println("TopicStringDataSubscribe string converted")
      }
    }
  }(stringReturnListener)
  return stringReturnListener
}

// pushes data to topic stack
func TopicPublishGlobData(clientStruct client, topicName string, data map[string]string) {
  encodedData := []byte(rcf_util.GlobMapEncode(data).Bytes())
  TopicPublishRawData(clientStruct, topicName, encodedData)
  InfoLogger.Println("TopicPublishGlobData called")
}

// pulls x msgs from topic topic stack
func TopicPullGlobData(clientStruct client, nmsgs int, topicName string) []map[string]string {
  InfoLogger.Println("TopicPullGlobData called")
  globMap := make([]map[string]string, 0)
  payloadMsgs := TopicPullRawData(clientStruct, topicName, nmsgs)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      globMap = append(globMap, rcf_util.GlobMapDecode(payloadMsg))
      InfoLogger.Println("TopicPullGlobData glob map converted")
    }
  }

  return globMap
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicGlobDataSubscribe(clientStruct client, topicName string) <-chan map[string]string{
  InfoLogger.Println("TopicGlobDataSubscribe called")
  rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
  stringReturnListener := make(chan map[string]string)
  go func(stringReturnListener chan<- map[string]string ){
    for {
      select {
        case rawData := <-rawReturnListener:
          stringReturnListener <- rcf_util.GlobMapDecode(rawData)
          InfoLogger.Println("TopicGlobDataSubscribe glob map converted")
      }
    }
  }(stringReturnListener)
  return stringReturnListener
}

//  executes action
func ActionExec(clientStruct client, actionName string, params []byte) {
  sendSlice := append(append([]byte(">action-"+actionName+"-exec-"), params...), "\r"...)
  clientStruct.Conn.Write(sendSlice)
  InfoLogger.Println("ActionExec called")
}

//  creates new action on node
func TopicCreate(clientStruct client, topicName string) {
  clientStruct.Conn.Write([]byte(">topic-"+topicName+"-create-\r"))
  InfoLogger.Println("TopicCreate called")
}

// lists node's topics
func TopicList(clientStruct client, connChannel chan []byte) []string {
  clientStruct.Conn.Write([]byte(">topic-all-list-\r"))
  data := <-connChannel
  InfoLogger.Println("TopicList called")
  return strings.Split(string(data), ",")
}
