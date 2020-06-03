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
  string Name
  string op

  chan LiveData
  chan BufferedData
}

var tcpConnBuffer = 1024
var topicContextMsgs chan []byte
var serviceContextMsgs chan []byte

var topicContextRequests []dataRequest
var serviceContextRequests []dataRequest

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
        ptype, name, operation, payload := rcf_util.ParseNodeReadProtocol(data) 
        if ptype != "" {
          if ptype == "topic" {
              topicContextMsgs <- data
            } else if ptype == "service" {
              serviceContextMsgs <- data
          }
        }
      }
    }
  }
}

func topicHandler(conn net.Conn, topicContextMsgs chan []byte, []topicContextRequests) {
  // client read protocol ><type>-<name>-<operation>-<paypload(msgs)>
  for {
    select {
      case data := <-topicContextRequests:
        //only for parsing purposes
        dataString := string(data)
        if strings.SplitN(dataString, "-", 4)[2] == "pull" {
          for _, request := range topicContextRequests {
            if request.BufferedData == nil && request.op == "pull" {
              payloadMsgs := TopicParsePulledRawData(data, request.name)
              if payloadMsgs =! nil {
                request.BufferedData = payloadMsgs  
              }
            }
          }
        } else if strings.SplitN(dataString, "-", 4)[2] == "sub" {
          for _, request := range topicContextRequests {
            if request.LiveData == nil && request.op == "sub" {
              var payload []byte
              if strings.Split(dataString, "-")[1] == request.Name {
                payload = bytes.SplitN(data, []byte("-"), 4)[3]
                request.LiveData <- payload
              }
            }
          }
        }
    }
  }
}

func serviceHandler(conn net.Conn, serviceContextMsgs chan []byte, []serviceContextRequests) {
  for {
    select {
      case data := <-serviceContextMsgs:
        for _, request := range serviceContextRequests {
          if request.LiveData == nil {
            payload := ParseServiceReplyPayload(data, request.name)
            if payload != nil {
              request.LiveData <- payload
            }
          }
        }
  }
}

func ParseServiceReplyPayload(data []byte, name string) []byte {  
  var payload []byte
  dataString := string(data)
  splitData := strings.Split(dataString, "-")
  
  if len(splitData) >= 2 {
    msgServiceName := splitData[1]
    if msgServiceName == name {
      payload = bytes.SplitN(data, []byte("-"), 4)[3]
    }
  }
  return payload
}

// pulls x msgs from topic topic stack
func TopicParsePulledRawData(data []byte, name string) [][]byte {
  var msgs [][]bytes

  if if strings.SplitN(dataString, "-", 4)[1] == name {
    msgTopicName := strings.Split(dataString, "-")[1]
    if msgTopicName == name {
      payload = bytes.SplitN(data, []byte("-"), 4)[3]
    }

    split_payload := bytes.Split(payload, []byte("\nm"))
    for _, splitPayloadMsg := range split_payload {
      if len(splitPayloadMsg) >= 1 {
        msgs = append(msgs, splitPayloadMsg)
      }
    }
  }
  return msgs
}

// function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed 
func connectToTcpServer(port int) (chan []byte, net.Conn) {
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))
  topicContextMsgs = make(chan []byte)
  serviceContextMsgs = make(chan []byte)
  topicContextMsgs = make([]topicContextRequests)
  serviceContextMsgs = make([]serviceContextRequests)

  go connHandler(conn, topicContextMsgs, serviceContextMsgs)
  
  go topicHandler(conn, topicContextMsgs, topicContextRequests)
  go serviceHandler(conn, serviceContextMsgs, serviceContextRequests)

  if err != nil {
    fmt.Println(err)
  }
  // don't forget to close connection
  return connChannel, conn
}

// returns connection to node
func NodeOpenConn(node_id int) (chan []byte, net.Conn) {
  return connectToTcpServer(node_id)
}

func NodeCloseConn(conn net.Conn) {
  conn.Write([]byte("end\r"))
  conn.Close()
}


// executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func ServiceExec(conn net.ConnserviceName string, params []byte) string {
  serviceListener := make(chan []byte)
  serviceId := rand.Intn(255)
  if serviceId == 0 || serviceId == 2 {
    serviceId = rand.Intn(255)  
  }
  name := serviceName+","+strconv.Itoa(serviceId)
  conn.Write(append(append([]byte(">service-"+name+"-exec-"), params...), "\r"...))

    
  return name
}

func TopicPullRawData(conn net.Conn, nmsgs int) sting {
  pullReqId := rand.Intn(255)
  if pullReqId == 0 || pullReqId == 2 {
    pullReqId = rand.Intn(255)  
  }
  name := topicName+","+strconv.Itoa(pullReqId)
  send_slice := append([]byte(">topic-"+name+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  conn.Write(send_slice)
  return name
}

func TopicRawDataSubscribe(net.Conn, topicName string) string {
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  return topicName
}
// pushes raw byte slice msg to topic msg stack
func TopicPublishRawData(conn net.Conn, topicName string, data []byte) {
  send_slice := append(append([]byte(">topic-"+topicName+"-publish-"),data...),"\r"...)
  conn.Write(send_slice)
}


// pushes string msg to topic msg stack
func TopicPublishStringData(conn net.Conn, topicName string, data string) {
  TopicPublishRawData(conn, topicName, []byte(data))
}

// pulls x msgs from topic topic stack
func TopicPullStringData(conn net.Conn, connChannel chan []byte, nmsgs int, topicName string) []string {
  var msgs []string
  payloadMsgs := TopicPullRawData(conn, connChannel, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      msgs = append(msgs, string(payloadMsg))
    }
  }

  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicStringDataSubscribe(conn net.Conn, connChannel chan []byte, topicName string) <-chan string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan string)
  go func(topicListener chan<- string){
    for {
      select {
        case data := <-connChannel:
          if len(data)>=1 {
            topicListener <- string(rcf_util.TopicParseClientReadPayload(data, topicName))
          }
      }
    }
  }(topicListener)
  return topicListener
}

// pushes data to topic stack
func TopicPublishGlobData(conn net.Conn, topicName string, data map[string]string) {
  encodedData := []byte(rcf_util.GlobMapEncode(data).Bytes())
  // println("Published data:")
  // fmt.Printf("%08b", encodedData)
  TopicPublishRawData(conn, topicName, encodedData)
}

// pulls x msgs from topic topic stack
func TopicPullGlobData(conn net.Conn, connChannel chan []byte, nmsgs int, topicName string) []map[string]string {
  glob_map := make([]map[string]string, 0)
  payloadMsgs := TopicPullRawData(conn, connChannel, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      glob_map = append(glob_map, rcf_util.GlobMapDecode(payloadMsg))
    }
  }
  return glob_map
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicGlobDataSubscribe(conn net.Conn, connChannel chan []byte, topicName string) <-chan map[string]string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan map[string]string)
  go func(topicListener chan<- map[string]string ){
    for {
      select {
        case sdata := <-connChannel:
          if len(sdata) >= 1 {
            payload := rcf_util.TopicParseClientReadPayload(sdata, topicName)
            if len(payload) >= 1 {
              data_map := rcf_util.GlobMapDecode(payload)
              topicListener <- data_map
            }
          }
      }
    }
  }(topicListener)
  return topicListener
}

//  executes action
func ActionExec(conn net.Conn, actionName string, params []byte) {
  send_slice := append(append([]byte(">action-"+actionName+"-exec-"), params...), "\r"...)
  conn.Write(send_slice)
}

//  creates new action on node
func TopicCreate(conn net.Conn, topicName string) {
  conn.Write([]byte(">topic-"+topicName+"-create-\r"))
}

// lists node's topics
func TopicList(conn net.Conn, connChannel chan []byte) []string {
  conn.Write([]byte(">topic-all-list-\r"))
  data := <-connChannel
  
  return strings.Split(string(data), ",")
}
