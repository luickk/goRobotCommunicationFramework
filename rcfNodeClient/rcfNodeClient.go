/*
Package rcfnodeclient implements all functions to communicate  with a rcf node.
*/
package rcfnodeclient

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	rcfUtil "rcf/rcfUtil"
	"strconv"
	"strings"
)

// struct that contains all information for a valid request for the handler which communicates with the give node
type dataRequest struct {
	// name always includes optional id
	Name string
	// or action
	Op string
	// true if request has been processed
	Fulfilled bool

	// the result of the request
	ReturnedPayload chan []byte
	// a pull operation requires a 2 dim slice, since it can contain multiple msgs
	PullOpReturnedPayload chan [][]byte
}

// client struct contains the connection to the node for write access and the request channels which are read/ processed by the handlers
type client struct {
	// connection to the node
	Conn net.Conn

	// topic pull/ sub request channel which is read/ processed by the topic handler
	TopicContextRequests chan *dataRequest
	// service call request channel which is read/ processed by the serice handler
	ServiceContextRequests chan *dataRequest
}

// buffer size for the connection handler which reads incoming information from the tcp socket
var tcpConnBuffer = 1024

// channel wich raw msgs from the node are pushed to if their type/ context is topic
var topicContextMsgs chan []byte

// channel wich raw msgs from the node are pushed to if their type/ context is service
var serviceContextMsgs chan []byte

// maximum capacity for active service, topic requests
var requestCapacity = 1000

// basic logger declarations
var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

// parses incoming instructions from the node and sorts them according to their context/ type
// pushes sorted instructions to the according handler
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
			if len(data) >= 1 {
				ptype, _, _, _, _ := rcfUtil.ParseNodeReadProtocol(data)
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

// handles topic pull/ sub requests and processes topic context/ type msg payloads
func topicHandler(conn net.Conn, topicContextMsgs chan []byte, topicRequests chan *dataRequest) {
	InfoLogger.Println("topicHandler started")
	requests := make([]dataRequest, requestCapacity)
	for {
		select {
		case data := <-topicContextMsgs:
			//only for parsing purposes
			dataString := string(data)
			operation := strings.SplitN(dataString, "-", 5)[2]
			if operation == "pull" {
				for i, request := range requests {
					if request.Op == "pull" && request.Fulfilled == false {
						InfoLogger.Println("topicHandler pull handled")
						payloadMsgs, dataValid := ParseTopicPulledRawData(data, request.Name)
						if dataValid {
							request.PullOpReturnedPayload <- payloadMsgs
							requests[i].Fulfilled = true
						} else {
							requests[i].Fulfilled = true
						}
					}
				}
			} else if operation == "sub" {
				for _, request := range requests {
					if request.Op == "sub" && request.Fulfilled == false {
						var payload []byte
						if strings.Split(dataString, "-")[1] == request.Name {
							InfoLogger.Println("topicHandler subscribed handled")
							if len(bytes.SplitN(data, []byte("-"), 5)) > 4 {
								payload = bytes.SplitN(data, []byte("-"), 5)[4]
								payloadLen, err := strconv.Atoi(strings.SplitN(dataString, "-", 5)[3])
								if err != nil {
									WarningLogger.Println("topicHandler subscribe payload len conversion error")
								} else {
									if len(payload) != payloadLen {
										WarningLogger.Println("topicHandler subscribe parsing payload extraction error")
									} else {
										request.ReturnedPayload <- payload
									}
								}
							} else {
								WarningLogger.Println("topicHandler subscribe protocol parsing err")
							}
						}
					}
				}
			} else if operation == "pullinfo" {
				for i, request := range requests {
					if request.Op == "pulltopiclist" && request.Fulfilled == false {
						var payload []byte
						InfoLogger.Println("topicHandler info topicList handled")
						if len(bytes.SplitN(data, []byte("-"), 5)) > 4 {
							payload = bytes.SplitN(data, []byte("-"), 5)[4]
							payloadLen, err := strconv.Atoi(strings.SplitN(dataString, "-", 5)[3])
							if err != nil {
								WarningLogger.Println("topicHandler info topicList payload len conversion error")
								requests[i].Fulfilled = true
							} else {
								if len(payload) != payloadLen {
									WarningLogger.Println("topicHandler info topicList parsing payload extraction error")
									requests[i].Fulfilled = true
								} else {
									request.ReturnedPayload <- payload
									requests[i].Fulfilled = true
								}
							}
						} else {
							WarningLogger.Println("topicHandler info topicList protocol parsing err")
							requests[i].Fulfilled = true
						}
					}
				}
			}
		case request := <-topicRequests:
			InfoLogger.Println("topicHandler request added")
			if len(requests) > requestCapacity {
				requestOverhead := len(requests) - requestCapacity
				// slicing size of slice to right size‚
				requests = requests[requestOverhead:]
			}
			requests = append(requests, *request)
		}
	}
}

// handles service call requests and processes the results which are contained in the service type/context msg payloads
func serviceHandler(conn net.Conn, serviceContextMsgs <-chan []byte, serviceRequests <-chan *dataRequest) {
	InfoLogger.Println("serviceHandler started")
	requests := make([]*dataRequest, requestCapacity)
	for {
		select {
		case data := <-serviceContextMsgs:
			if len(data) >= 0 {
				for i, request := range requests {
					if request == nil {
						break
					}
					if request.Fulfilled == false && request.Name != "" {
						InfoLogger.Println("serviceHandler service done executing")
						payload, dataValid := ParseServiceReplyPayload(data, request.Name)
						if dataValid {
							if len(payload) != 0 {
								InfoLogger.Println("serviceHandler service payload returned")
								request.ReturnedPayload <- payload
								requests[i].Fulfilled = true
							}
						} else {
							requests[i].Fulfilled = true
						}
					}
				}
			}
		case request := <-serviceRequests:
			InfoLogger.Println("serviceHandler request added")
			if len(requests) > requestCapacity {
				requestOverhead := len(requests) - requestCapacity
				// slicing size of slice to right size‚
				requests = requests[requestOverhead:]
			}
			requests = append(requests, request)
		}
	}
}

// ServiceExec executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func ServiceExec(clientStruct client, serviceName string, params []byte) []byte {
	InfoLogger.Println("ServiceExec service exec called")
	serviceID := rcfUtil.GenRandomIntID()
	name := serviceName + "," + strconv.Itoa(serviceID)

	request := new(dataRequest)
	request.Name = name
	request.Op = "exec"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	clientStruct.ServiceContextRequests <- request
	clientStruct.Conn.Write(append(append([]byte(">service-"+name+"-exec-"+strconv.Itoa(len(params))+"-"), params...), "\r"...))
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

// TopicPullRawData Pulls raw data msgs from given topic
func TopicPullRawData(clientStruct client, topicName string, nmsgs int) [][]byte {
	InfoLogger.Println("TopicPullRawData called")
	// generates random id for the name
	pullReqID := rcfUtil.GenRandomIntID()
	name := topicName + "," + strconv.Itoa(pullReqID)
	// create instrucitons slice for the node according to the protocl
	instructionSlice := append([]byte(">topic-"+name+"-pull-"+strconv.Itoa(nmsgs)+"-"), "\r"...)
	clientStruct.Conn.Write(instructionSlice)

	// creating request for the payload which is sent back from the node
	request := new(dataRequest)
	request.Name = name
	request.Op = "pull"
	request.Fulfilled = false
	request.PullOpReturnedPayload = make(chan [][]byte)
	// pushing request to topic handler where it is process
	clientStruct.TopicContextRequests <- request
	InfoLogger.Println("TopicPullRawData request sent")

	reply := false
	payload := [][]byte{}

	// wainting for request to be processed and retrieval of payload
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

// TopicRawDataSubscribe subscribes to topic and pulls raw msgs data
func TopicRawDataSubscribe(clientStruct client, topicName string) chan []byte {
	InfoLogger.Println("TopicRawDataSubscribe called")
	// generating random id for the name
	pullReqID := rcfUtil.GenRandomIntID()
	name := topicName + "," + strconv.Itoa(pullReqID)
	// creating and writing instruction slice for the node
	clientStruct.Conn.Write([]byte(">topic-" + name + "-subscribe-0-\r"))

	// creating request for topic handler
	request := new(dataRequest)
	request.Name = name
	request.Op = "sub"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	// sending request to topic handler
	clientStruct.TopicContextRequests <- request
	InfoLogger.Println("TopicRawDataSubscribe request sent and channel returned")

	// returning channel from request to which the topic handler writes the results
	return request.ReturnedPayload
}

// connectToTCPServer function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed
func connectToTCPServer(port int) (net.Conn, chan []byte, chan []byte) {
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

// NodeOpenConn initiates loggers and comm channels for handler and start handlers
// returns client struct which defines relevant information for the interface functions to work
func NodeOpenConn(nodeID int) client {
	InfoLogger = log.New(os.Stdout, "[CLIENT] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[CLIENT] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[CLIENT] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.SetOutput(ioutil.Discard)

	rcfUtil.InfoLogger = InfoLogger
	rcfUtil.WarningLogger = WarningLogger
	rcfUtil.ErrorLogger = ErrorLogger

	InfoLogger.Println("NodeOpenConn called")
	conn, topicContextMsgs, serviceContextMsgs := connectToTCPServer(nodeID)
	topicContextRequests := make(chan *dataRequest)
	serviceContextRequests := make(chan *dataRequest)

	go topicHandler(conn, topicContextMsgs, topicContextRequests)
	go serviceHandler(conn, serviceContextMsgs, serviceContextRequests)
	InfoLogger.Println("Init handler routines started")

	client := new(client)
	client.Conn = conn
	client.TopicContextRequests = topicContextRequests
	client.ServiceContextRequests = serviceContextRequests

	return *client
}

// NodeCloseConn closes node conn
func NodeCloseConn(clientStruct client) {
	InfoLogger.Println("NodeCloseConn closed")
	clientStruct.Conn.Write([]byte("end\r"))
	clientStruct.Conn.Close()
}

// ParseServiceReplyPayload parses the payload from service type/ context msgs according to the protocl
// returns payload and a bool wther instructions were valid or not
func ParseServiceReplyPayload(data []byte, name string) ([]byte, bool) {
	InfoLogger.Println("ParseServiceReplyPayload called")
	couldBeParsed := true
	payload := []byte{}
	dataString := string(data)
	splitData := strings.SplitN(dataString, "-", 5)
	msgTopicName := splitData[1]
	finalPayload := []byte{}

	// checks if instruction is valid
	if name != "" {
		if msgTopicName == name {
			if len(splitData) > 4 {
				payload = bytes.SplitN(data, []byte("-"), 5)[4]
				payloadLen, err := strconv.Atoi(strings.SplitN(dataString, "-", 5)[3])
				if err != nil {
					WarningLogger.Println(dataString)
					WarningLogger.Println("ParseServiceReplyPayload payload len conversion error")
					couldBeParsed = false
				} else {
					if len(payload) != payloadLen {
						WarningLogger.Println("ParseServiceReplyPayload parsing payload extraction error")
						couldBeParsed = false
					} else {
						finalPayload = payload
					}
				}
			} else {
				WarningLogger.Println("ParseServiceReplyPayload protocol parsing err")
				couldBeParsed = false
			}
		}
	} else {
		WarningLogger.Println("serviceHandler missing request name attr")
		couldBeParsed = false
	}
	return finalPayload, couldBeParsed
}

// ParseTopicPulledRawData parses the payload from topic type/ context msgs according to the protocol
// returns payload and a bool wther instructions were valid or not
func ParseTopicPulledRawData(data []byte, name string) ([][]byte, bool) {
	InfoLogger.Println("ParseTopicPulledRawData called")
	couldBeParsed := true
	msgs := [][]byte{}
	payload := []byte{}
	dataString := string(data)
	msgTopicName := strings.SplitN(dataString, "-", 5)[1]
	// checks if given request name equals the name parsed from the msg
	if name != "" {
		if msgTopicName == name {
			// splits the payload from the instruction
			payload = bytes.SplitN(data, []byte("-"), 5)[4]
			payloadStringLengths := strings.Split(string(bytes.SplitN(data, []byte("-"), 5)[3]), ",")
			payloadIntLengths := make([]int, len(payloadStringLengths))
			for i, s := range payloadStringLengths {
				conv, err := strconv.Atoi(s)
				if err != nil {
					WarningLogger.Println("ParseTopicPulledRawData payload len conversion error")
				} else {
					payloadIntLengths[i] = conv
				}
			}

			splitPayload := [][]byte{}
			payloadLastSplit := 0
			for _, length := range payloadIntLengths {
				splitPayload = append(splitPayload, payload[payloadLastSplit:payloadLastSplit+length])
				payloadLastSplit += length
			}

			// iterates over split msgs and appends them to result slice
			for _, splitPayloadMsg := range splitPayload {
				if len(splitPayloadMsg) >= 1 {
					InfoLogger.Println("ParseTopicPulledRawData payload returned")
					msgs = append(msgs, splitPayloadMsg)
				}
			}
		}
	} else {
		couldBeParsed = false
		WarningLogger.Println("serviceHandler missing request name attr")
	}
	return msgs, couldBeParsed
}

// TopicPublishRawData pushes raw byte slice msg to topic msg stack
func TopicPublishRawData(clientStruct client, topicName string, data []byte) {
	InfoLogger.Println("TopicPublishRawData called")
	sendSlice := append(append([]byte(">topic-"+topicName+"-publish-"+strconv.Itoa(len(data))+"-"), data...), "\r"...)
	clientStruct.Conn.Write(sendSlice)
}

// TopicPublishStringData pushes string msg to topic msg stack
func TopicPublishStringData(clientStruct client, topicName string, data string) {
	InfoLogger.Println("TopicPublishStringData called")
	TopicPublishRawData(clientStruct, topicName, []byte(data))
}

// TopicPullStringData pulls x msgs from topic topic stack and descodes them as string
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

// TopicStringDataSubscribe waits for incoming topic msgs on subscribed channel
// returns the string encoded msgs
func TopicStringDataSubscribe(clientStruct client, topicName string) <-chan string {
	InfoLogger.Println("TopicStringDataSubscribe called")
	rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
	stringReturnListener := make(chan string)
	go func(stringReturnListener chan<- string) {
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

// TopicPublishGlobData pushes data to topic stack
func TopicPublishGlobData(clientStruct client, topicName string, data map[string]string) {
	InfoLogger.Println("TopicPublishGlobData called")
	encodedData, err := rcfUtil.GlobMapEncode(data)
	encodedDataSlice := []byte(encodedData.Bytes())
	if err != nil {
		WarningLogger.Println("GlobMapEncode encoding error")
		WarningLogger.Println(err)
	} else {
		TopicPublishRawData(clientStruct, topicName, encodedDataSlice)
	}
}

// TopicPullGlobData pulls x msgs from topic topic stack
func TopicPullGlobData(clientStruct client, nmsgs int, topicName string) []map[string]string {
	InfoLogger.Println("TopicPullGlobData called")
	globMap := make([]map[string]string, 0)
	payloadMsgs := TopicPullRawData(clientStruct, topicName, nmsgs)
	for _, payloadMsg := range payloadMsgs {
		if len(payloadMsg) > 1 {
			pulld, err := rcfUtil.GlobMapDecode(payloadMsg, "pull")
			if err == nil {
				globMap = append(globMap, pulld)
				InfoLogger.Println("TopicPullGlobData glob map converted")
			} else {
				InfoLogger.Println("TopicPullGlobData glob map conversion failed!")
				WarningLogger.Println("TopicPullGlobData glob map conversion failed!")
			}
		}
	}

	return globMap
}

// TopicGlobDataSubscribe waits for incoming topic msgs on subscribed channel
// returns the glob encoded msgs
func TopicGlobDataSubscribe(clientStruct client, topicName string) <-chan map[string]string {
	InfoLogger.Println("TopicGlobDataSubscribe called")
	rawReturnListener := TopicRawDataSubscribe(clientStruct, topicName)
	stringReturnListener := make(chan map[string]string)
	go func(stringReturnListener chan<- map[string]string) {
		for {
			select {
			case rawData := <-rawReturnListener:
				if len(rawData) > 1 {
					pulld, err := rcfUtil.GlobMapDecode(rawData, "subs")
					if err == nil {
						stringReturnListener <- pulld
						InfoLogger.Println("TopicGlobDataSubscribe glob map converted")
					} else {
						WarningLogger.Println("TopicGlobDataSubscribe glob map conversion failed!")
					}
				}
			}
		}
	}(stringReturnListener)
	return stringReturnListener
}

// ActionExec executes action
func ActionExec(clientStruct client, actionName string, params []byte) {
	InfoLogger.Println("ActionExec called")
	sendSlice := append(append([]byte(">action-"+actionName+"-exec-"+strconv.Itoa(len(params))+"-"), params...), "\r"...)
	clientStruct.Conn.Write(sendSlice)
}

// TopicCreate creates new action on node
func TopicCreate(clientStruct client, topicName string) {
	InfoLogger.Println("TopicCreate called")
	clientStruct.Conn.Write([]byte(">topic-" + topicName + "-create-0-\r"))
}

// TopicList lists node's topics
func TopicList(clientStruct client) []string {
	InfoLogger.Println("TopicList called")
	clientStruct.Conn.Write([]byte(">topic-all-list-0-\r"))

	// creating request for the payload which is sent back from the node
	request := new(dataRequest)
	request.Name = "topiclist"
	request.Op = "pulltopiclist"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	// pushing request to topic handler where it is process
	clientStruct.TopicContextRequests <- request
	InfoLogger.Println("TopicList request sent")

	reply := false
	payload := []byte{}

	// wainting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.ReturnedPayload:
			payload = liveDataRes
			InfoLogger.Println("TopicList payload returned")
			reply = true
			break
		}
	}
	return strings.Split(string(payload), ",")
}
