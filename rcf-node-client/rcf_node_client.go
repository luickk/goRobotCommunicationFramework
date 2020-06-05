package rcf_node_client

import(
	"fmt"
	"net"
	"strconv"
	"strings"
	"bufio"
	"bytes"
  "math/rand"
	"rcf/rcf-util"
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

// parses information and pushes them to byte channel 
func connHandler(conn net.Conn, topicContextMsgs chan []byte, serviceContextMsgs chan []byte) {
  for {
    data := make([]byte, tcpConnBuffer)
    // println("received data: "+string(data))
    n, _ := bufio.NewReader(conn).Read(data)
    data = data[:n]
    splitRData := bytes.Split(data, []byte("\r"))
    for _, data := range splitRData {
      if len(data)>=1 {
        // data is parsed according to the node read protocol
        // ><type>-<name>-<operation>-<paypload byte slice>
        ptype, _, _, _ := rcf_util.ParseNodeReadProtocol(data) 
        println(ptype)
        if ptype != "" {
          if ptype == "topic" {
              topicContextMsgs <- data
            } else if ptype == "service" {
              println("service type")
              serviceContextMsgs <- data
          }
        }
      }
    }
  }
}

func topicHandler(conn net.Conn, topicContextMsgs chan []byte, topicRequests chan dataRequest) {
  requests := []dataRequest{}
  for {
    select {
      case data := <-topicContextMsgs:
        //only for parsing purposes
        dataString := string(data)
        if strings.SplitN(dataString, "-", 4)[2] == "pull"{
          for i, request := range requests {
            if request.Op == "pull" && request.Fulfilled == false {
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
                payload = bytes.SplitN(data, []byte("-"), 4)[3]
                request.ReturnedPayload <- payload
                // requests[i].Fulfilled = true
              }
            }
          }
        }
      case request := <-topicRequests:
        requests = append(requests, request)
    }
  }
}

func serviceHandler(conn net.Conn, serviceContextMsgs <-chan []byte, serviceRequests <-chan dataRequest) {
  requests := []dataRequest{}
  for {
    select {
      case data := <-serviceContextMsgs:
        for i, request := range requests {
          if request.Fulfilled == false {
            println("parsing")
            payload := ParseServiceReplyPayload(data, request.Name)
            if len(payload) != 0 {
              request.ReturnedPayload <- payload
              requests[i].Fulfilled = true
            }
          }
        }
      case request := <-serviceRequests:
        println("request added")
        requests = append(requests, request)
    }
  }
}

// executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func ServiceExec(clientStruct client, serviceName string, params []byte) []byte {
  serviceId := rand.Intn(255)
  if serviceId == 0 || serviceId == 2 {
    serviceId = rand.Intn(255)  
  }
  name := serviceName+","+strconv.Itoa(serviceId)

  request := new(dataRequest)
  request.Name = name
  request.Op = "exec"
  request.Fulfilled = false
  request.ReturnedPayload = make(chan []byte)
  clientStruct.ServiceContextRequests <- *request
  clientStruct.Conn.Write(append(append([]byte(">service-"+name+"-exec-"), params...), "\r"...))
  
  reply := false
  payload := []byte{}

  for !reply {
    select {
      case liveDataRes := <-request.ReturnedPayload:
        payload = liveDataRes
        close(request.ReturnedPayload)
        break
        reply = true      
    }
  }

  return payload
}

func TopicPullRawData(clientStruct client, topicName string, nmsgs int) [][]byte {
  pullReqId := rand.Intn(255) 
  if pullReqId == 0 || pullReqId == 2 {
    pullReqId = rand.Intn(255)  
  }
  name := topicName+","+strconv.Itoa(pullReqId)
  instructionSlice := append([]byte(">topic-"+name+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  clientStruct.Conn.Write(instructionSlice)


  request := new(dataRequest)
  request.Name = name
  request.Op = "pull"
  request.Fulfilled = false
  request.PullOpReturnedPayload = make(chan [][]byte)
  clientStruct.TopicContextRequests <- *request

  reply := false
  payload := [][]byte{}

  for !reply {
    select {
      case liveDataRes := <-request.PullOpReturnedPayload:
        payload = liveDataRes
        reply = true      
        break
    }
  }
  return payload
}

func TopicRawDataSubscribe(clientStruct client, topicName string) chan []byte {
  clientStruct.Conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))

  request := new(dataRequest)
  request.Name = topicName
  request.Op = "sub"
  request.Fulfilled = false
  request.ReturnedPayload = make(chan []byte)
  clientStruct.TopicContextRequests <- *request

  return request.ReturnedPayload
}


// function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed 
func connectToTcpServer(port int) (net.Conn, chan []byte, chan []byte) {
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))
  topicContextMsgs = make(chan []byte)
  serviceContextMsgs = make(chan []byte)

  go connHandler(conn, topicContextMsgs, serviceContextMsgs)

  if err != nil {
    fmt.Println(err)
  }
  // don't forget to close connection
  return conn, topicContextMsgs, serviceContextMsgs
}

// returns connection to node
func NodeOpenConn(node_id int) client {
  conn, topicContextMsgs, serviceContextMsgs := connectToTcpServer(node_id)
  topicContextRequests := make(chan dataRequest)
  serviceContextRequests := make(chan dataRequest)

  
  go topicHandler(conn, topicContextMsgs, topicContextRequests)
  go serviceHandler(conn, serviceContextMsgs, serviceContextRequests)

  client := new(client)
  client.Conn = conn
  client.TopicContextRequests = topicContextRequests
  client.ServiceContextRequests = serviceContextRequests
  
  return *client
}

func NodeCloseConn(clientStruct client) {
  clientStruct.Conn.Write([]byte("end\r"))
  clientStruct.Conn.Close()
}

func ParseServiceReplyPayload(data []byte, name string) []byte {  
  var payload []byte
  dataString := string(data)
  splitData := strings.Split(dataString, "-")
  
  if len(splitData) >= 2 {
    msgServiceName := splitData[1]
    println(msgServiceName)
    println(name)
    if msgServiceName == name {
      payload = bytes.SplitN(data, []byte("-"), 4)[3]
      println("payload")
    }
  }
  return payload
}

// pulls x msgs from topic topic stack
func ParseTopicPulledRawData(data []byte, name string) [][]byte {
  var msgs [][]byte
  var payload []byte
  dataString := string(data)
  msgTopicName := strings.SplitN(dataString, "-", 4)[1]
  if msgTopicName == name {
    payload = bytes.SplitN(data, []byte("-"), 4)[3]
    splitPayload := bytes.Split(payload, []byte("\nm"))

    for _, splitPayloadMsg := range splitPayload {
      if len(splitPayloadMsg) >= 1 {
        msgs = append(msgs, splitPayloadMsg)
      }
    }
  }
  return msgs
}

// pushes raw byte slice msg to topic msg stack
func TopicPublishRawData(clientStruct client, topicName string, data []byte) {
  send_slice := append(append([]byte(">topic-"+topicName+"-publish-"),data...),"\r"...)
  clientStruct.Conn.Write(send_slice)
}


// pushes string msg to topic msg stack
func TopicPublishStringData(clientStruct client, topicName string, data string) {
  TopicPublishRawData(clientStruct, topicName, []byte(data))
}

// pulls x msgs from topic topic stack
func TopicPullStringData(clientStruct client, nmsgs int, topicName string) []string {
  var stringPayload []string
  payloadMsgs := TopicPullRawData(clientStruct, topicName, nmsgs)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      stringPayload = append(stringPayload, string(payloadMsg))
    }
  }

  return stringPayload
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicStringDataSubscribe(clientStruct client, topicName string) <-chan string {
  rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
  stringReturnListener := make(chan string)
  go func(stringReturnListener chan<- string ){
    for {
      select {
        case rawData := <-rawReturnListener:
          stringReturnListener <- string(rawData)
      }
    }
  }(stringReturnListener)
  return stringReturnListener
}

// pushes data to topic stack
func TopicPublishGlobData(clientStruct client, topicName string, data map[string]string) {
  encodedData := []byte(rcf_util.GlobMapEncode(data).Bytes())
  // println("Published data:")
  // fmt.Printf("%08b", encodedData)
  TopicPublishRawData(clientStruct, topicName, encodedData)
}

// pulls x msgs from topic topic stack
func TopicPullGlobData(clientStruct client, nmsgs int, topicName string) []map[string]string {
  globMap := make([]map[string]string, 0)
  payloadMsgs := TopicPullRawData(clientStruct, topicName, nmsgs)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      globMap = append(globMap, rcf_util.GlobMapDecode(payloadMsg))
    }
  }

  return globMap
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicGlobDataSubscribe(clientStruct client, topicName string) <-chan map[string]string{
  rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
  stringReturnListener := make(chan map[string]string)
  go func(stringReturnListener chan<- map[string]string ){
    for {
      select {
        case rawData := <-rawReturnListener:
          stringReturnListener <- rcf_util.GlobMapDecode(rawData)
      }
    }
  }(stringReturnListener)
  return stringReturnListener
}

//  executes action
func ActionExec(clientStruct client, actionName string, params []byte) {
  send_slice := append(append([]byte(">action-"+actionName+"-exec-"), params...), "\r"...)
  clientStruct.Conn.Write(send_slice)
}

//  creates new action on node
func TopicCreate(clientStruct client, topicName string) {
  clientStruct.Conn.Write([]byte(">topic-"+topicName+"-create-\r"))
}

// lists node's topics
func TopicList(clientStruct client, connChannel chan []byte) []string {
  clientStruct.Conn.Write([]byte(">topic-all-list-\r"))
  data := <-connChannel
  
  return strings.Split(string(data), ",")
}
