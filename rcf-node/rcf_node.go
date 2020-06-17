/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_node

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
	rcf_util "rcf/rcf-util"
	"strconv"
	"strings"
	"time"
	"runtime"
	"io/ioutil"
)

// single topic msg capacity
// the amount of msgs a single topic "stores" before they get deleted
var topicCapacity = 5

// frequency with which nodes handlers are refreshed
var nodeFreq = 0

var tcpConnBuffer = 1024


// basic logger declarations
var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

// describes a single client write request
// exists to forward client write requests form the handlers to the clientWrite handler via a channel
type clientWriteRequest struct {
	receivingClient net.Conn
	msg []byte
}

// describes a single msgs
// exists to forward requests form the connection handler to the topic handler via a channel in pull,push/ subscribe events
type topicMsg struct {
	topicName string
	msg       []byte
}

// describes a pull request from a client
// exists to send requests form the connection handler to the topic handler via a channel
type topicPullReq struct {
	conn      net.Conn
	topicName string
	nmsg      int
}

// describes a single connection which is subscribed to a certain topic
// exists to forward subscribe requests form the connection handler to the topic handler via a channel
type topicListenerConn struct {
	listeningConn net.Conn
	topicName     string
}

// defines a single action with the executed function and its name
// exists to forward created actions from interface function to action handler
type action struct {
	actionName     string
	actionFunction actionFn
}

// defines a single service with the executed function and its name
// exists to forward created services from interface function to service handler
type service struct {
	serviceName     string
	serviceFunction serviceFn
}

// defines an services call request
// contains the connection it is called by such as the called services name and the call params
type serviceExec struct {
	serviceName     string
	serviceCallConn net.Conn
	params          []byte
}

// defines as action call request
// contains the connection it is called by such as the called action name and the call params
type actionExec struct {
	actionName string
	params     []byte
}

// node struct
type Node struct {
	// id or port of node
	id int

	// topic map is handled by the topic handler and contains all topics data
	// key: topic name, value: stack slice
	clientWriteRequestCh chan clientWriteRequest 

	// topic map is handled by the topic handler and contains all topics data
	// key: topic name, value: stack slice
	topics map[string][][]byte

	// contains topic msgs which should be pushed to a topic
	// handled by topic handler
	topicPushCh chan topicMsg

	// contains topics which should be created
	// handled by topic handler
	topicCreateCh chan string

	// contains topic subscribe requests which should be processed by the handler handler
	topicListenerConnCh chan topicListenerConn

	// contains all topic pull request which should be processed by the topic handler
	topicPullCh chan topicPullReq

	// contains all active topic listeners
	// handled by topic handler
	topicListenerConns []topicListenerConn

	// actions map is handled by the action handler and contains all actions
	// key: action name, value: action function
	actions map[string]actionFn

	// contains actions which should be created
	// handled by action handler
	actionCreateCh chan action

	// contains actions which should be executed
	// handled by action handler
	actionExecCh chan actionExec

	// services map is handled by the service handler and contains all services
	// key: service name, value: service function
	services map[string]serviceFn

	// cotains services which should be created
	// handled by service handler
	serviceCreateCh chan service

	// contains service s which should be executed
	// handled by service handler
	serviceExecCh chan serviceExec
}

// general action function type
type actionFn func(params []byte, nodeInstance Node)

// general service function type
type serviceFn func(params []byte, nodeInstance Node) []byte

// handles every incoming instructions from the client and executes them
func handleConnection(node Node, conn net.Conn) {
	InfoLogger.Println("handleConnection started")
	defer conn.Close()

	for {
		byteData := make([]byte, tcpConnBuffer)
		n, err := bufio.NewReader(conn).Read(byteData)
		if err != nil {
			ErrorLogger.Println("handleConnection routine reading error")
			ErrorLogger.Println(err)
			break
		}
		byteData = byteData[:n]
		// data := string(byteData)

		delimSplitByteData := bytes.Split(byteData, []byte("\r"))

		// iterating ovre conn read buffer array, split by backslash r
		for _, cmdByte := range delimSplitByteData {

			// parsing instrucitons from client
			ptype, name, operation, payload := rcf_util.ParseNodeReadProtocol(cmdByte)

			// cheks if instruction is valid
			if ptype != "" && name != "" {
				if ptype == "topic" {
					if operation == "publish" {
						InfoLogger.Println("handleConnection data published")
						// publishing data from client on given topic
						TopicPublishData(node, name, payload)
					} else if operation == "pull" {
						InfoLogger.Println("handleConnection data pulled")
						// handles pull reueqst from client
						nmsg, err := strconv.Atoi(string(payload))
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
						conn.Write(append([]byte(">info-list_topics-req-"), []byte(strings.Join(NodeListTopics(node), ",")+"\r")...))
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

// handles all write request to clients
func clientWriteRequestHandler(node Node) {
	InfoLogger.Println("writeHandler started")
	for {
		select {
			case writeRequest := <-node.clientWriteRequestCh:
				writeRequest.receivingClient.Write(writeRequest.msg)
			default:
		}
	}
}
// handles all memory critical read, write operations to the topics map and reduces the topic maps slices to given max length
// as well as pull,pus,sub operations from the client
func topicHandler(node Node) {
	InfoLogger.Println("topicHandler called")
	for {
		select {
			case topicListener := <-node.topicListenerConnCh:
				InfoLogger.Println("topicHandler listener added")
				// appends new client listener to active listener slice
				node.topicListenerConns = append(node.topicListenerConns, topicListener)
			case pullRequest := <-node.topicPullCh:
				InfoLogger.Println("topicHandler data pulled")
				var byteData [][]byte
				// parses the non unique topic name from the instruction
				topicOnlyName, _ := rcf_util.SplitServiceToNameId(pullRequest.topicName)

				if pullRequest.nmsg >= len(node.topics[topicOnlyName]) {
					byteData = node.topics[topicOnlyName]
				} else {
					byteData = node.topics[topicOnlyName][:pullRequest.nmsg]
				}

				if pullRequest.nmsg <= 1 {
					if len(byteData) >= 1 {
						// pullRequest.conn.Write(append(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), byteData[0]...), []byte("\r")...))
						clientWriteRequest := new(clientWriteRequest)
						clientWriteRequest.receivingClient = pullRequest.conn
						clientWriteRequest.msg = append(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), byteData[0]...), []byte("\r")...)
						node.clientWriteRequestCh <- *clientWriteRequest
					} else {
						// pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...))
						clientWriteRequest := new(clientWriteRequest)
						clientWriteRequest.receivingClient = pullRequest.conn
						clientWriteRequest.msg = append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...)
						node.clientWriteRequestCh <- *clientWriteRequest
					}
				} else {
					if len(byteData) >= 1 {
						tdata := append(bytes.Join(byteData, []byte("\nm")), []byte("\r")...)
						// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
						// pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), tdata...))
						clientWriteRequest := new(clientWriteRequest)
						clientWriteRequest.receivingClient = pullRequest.conn
						clientWriteRequest.msg = append([]byte(">topic-"+pullRequest.topicName+"-pull-"), tdata...)
						node.clientWriteRequestCh <- *clientWriteRequest
					} else {
						// pullRequest.conn.Write(append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...))
						clientWriteRequest := new(clientWriteRequest)
						clientWriteRequest.receivingClient = pullRequest.conn
						clientWriteRequest.msg = append([]byte(">topic-"+pullRequest.topicName+"-pull-"), []byte("\r")...)
						node.clientWriteRequestCh <- *clientWriteRequest
					}
				}
			case topicMsg := <-node.topicPushCh:
				InfoLogger.Println("topicHandler data pushed")

				if rcf_util.TopicsContainTopic(node.topics, topicMsg.topicName) {

					node.topics[topicMsg.topicName] = append(node.topics[topicMsg.topicName], topicMsg.msg)


					// check if topic exceeds topic cap limits
					if len(node.topics[topicMsg.topicName]) > topicCapacity {

						topicOverhead := len(node.topics[topicMsg.topicName]) - topicCapacity
						// slicing size of slice to right size‚
						node.topics[topicMsg.topicName] = node.topics[topicMsg.topicName][topicOverhead:]

					}

					// check if topic, which data is pushed to, has a listening conn
					for _, topicListener := range node.topicListenerConns {

						topicOnlyName, _ := rcf_util.SplitServiceToNameId(topicListener.topicName)

						if topicOnlyName == topicMsg.topicName {

							// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
							// topicListener.listeningConn.Write(append(append([]byte(">topic-"+topicListener.topicName+"-sub-"), []byte(topicMsg.msg)...), []byte("\r")...))
							clientWriteRequest := new(clientWriteRequest)
							clientWriteRequest.receivingClient = topicListener.listeningConn
							clientWriteRequest.msg = append(append([]byte(">topic-"+topicListener.topicName+"-sub-"), []byte(topicMsg.msg)...), []byte("\r")...)
							node.clientWriteRequestCh <- *clientWriteRequest
						}
					}
				}

			case topicCreateName := <-node.topicCreateCh:

				if !rcf_util.TopicsContainTopic(node.topics, topicCreateName) {
					node.topics[topicCreateName] = [][]byte{}
					InfoLogger.Println("topicHandler topic created")
				}
			}
			time.Sleep(time.Duration(nodeFreq))
	}
}

// handles all memory critical read, write operations to the actions map as well as create, execution instruction operations from the client
func actionHandler(nodeInstance Node) {
	InfoLogger.Println("actionHandler started")
	for {
		select {
		case action := <-nodeInstance.actionCreateCh:
			nodeInstance.actions[action.actionName] = action.actionFunction
			InfoLogger.Println("actionHandler action created")
		case actionExec := <-nodeInstance.actionExecCh:
			InfoLogger.Println("actionHandler action execed")
			if _, ok := nodeInstance.actions[actionExec.actionName]; ok {
				actionFunc := nodeInstance.actions[actionExec.actionName]
				go actionFunc(actionExec.params, nodeInstance)
			} else {
				InfoLogger.Println("actionHandler action execed not found")
			}
		}
		time.Sleep(time.Duration(nodeFreq))
	}
}

// handles all memory critical read, write operations to the services map as well as create, execution instruction operations from the client
func serviceHandler(nodeInstance Node) {
	InfoLogger.Println("actionHandler called")
	for {
		select {
		case service := <-nodeInstance.serviceCreateCh:
			nodeInstance.services[service.serviceName] = service.serviceFunction
			InfoLogger.Println("serviceHandler service created")
		case serviceExec := <-nodeInstance.serviceExecCh:
			InfoLogger.Println("serviceHandler service execed(queued)")
			serviceOnlyName, _ := rcf_util.SplitServiceToNameId(serviceExec.serviceName)
			if _, ok := nodeInstance.services[serviceOnlyName]; ok {
				go func() {
					serviceResult := append(nodeInstance.services[serviceOnlyName](serviceExec.params, nodeInstance), "\r"...)
					// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
					// serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), serviceResult...))
					clientWriteRequest := new(clientWriteRequest)
					clientWriteRequest.receivingClient = serviceExec.serviceCallConn
					clientWriteRequest.msg = append([]byte(">service-"+serviceExec.serviceName+"-called-"), serviceResult...)
					nodeInstance.clientWriteRequestCh <- *clientWriteRequest

					InfoLogger.Println("serviceHandler service returned and wrote payload")
				}()
			} else {
				// serviceExec.serviceCallConn.Write(append([]byte(">service-"+serviceExec.serviceName+"-called-"), []byte(serviceExec.serviceName+" not found \r")...))
				clientWriteRequest := new(clientWriteRequest)
				clientWriteRequest.receivingClient = serviceExec.serviceCallConn
				clientWriteRequest.msg = append([]byte(">service-"+serviceExec.serviceName+"-called-"), []byte(serviceExec.serviceName+" not found \r")...)
				nodeInstance.clientWriteRequestCh <- *clientWriteRequest
				InfoLogger.Println("serviceHandler service not found")
			}
			time.Sleep(time.Duration(nodeFreq))
		}
	}
}

// initiates node instance and initializes all channels, maps
func Create(nodeId int) Node {
	
	clientWriteRequestCh := make(chan clientWriteRequest)

	// key: topic name, value: stack slice
	topics := make(map[string][][]byte)

	topicPushCh := make(chan topicMsg)

	topicCreateCh := make(chan string)

	topicListenerConnCh := make(chan topicListenerConn)

	topicPullCh := make(chan topicPullReq)

	topicListenerConns := make([]topicListenerConn, 0)

	// action map with first key(action name) value(anon action func) pair
	actions := make(map[string]actionFn)

	actionCreateCh := make(chan action)

	actionExecCh := make(chan actionExec)

	services := make(map[string]serviceFn)

	serviceCreateCh := make(chan service)

	serviceExecCh := make(chan serviceExec)

	return Node{nodeId, clientWriteRequestCh, topics, topicPushCh, topicCreateCh, topicListenerConnCh, topicPullCh, topicListenerConns, actions, actionCreateCh, actionExecCh, services, serviceCreateCh, serviceExecCh}
}

// initiates node with given id
// returns initiated node instance to enable direct service and topic operations
func Init(node Node) {
	runtime.SetBlockProfileRate(1)
	InfoLogger = log.New(os.Stdout, "[NODE] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[NODE] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[NODE] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// disableing debug information
	InfoLogger.SetOutput(ioutil.Discard)
	ErrorLogger.SetOutput(ioutil.Discard)
	WarningLogger.SetOutput(ioutil.Discard)

	// initiating basic loggers
	rcf_util.InfoLogger = InfoLogger
	rcf_util.WarningLogger = WarningLogger
	rcf_util.ErrorLogger = ErrorLogger

	// starting all handlers
	go topicHandler(node)
	go actionHandler(node)
	go serviceHandler(node)
	go clientWriteRequestHandler(node)

	InfoLogger.Println("Init handlers routine started")

	var port string = ":" + strconv.Itoa(node.id)

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

// pauses node
func NodeHalt() {
	InfoLogger.Println("NodeHalt called")
	for {
		time.Sleep(1 * time.Second)
	}
}

// adds new subscribe request to active listener conn slice
func TopicAddListenerConn(node Node, topicName string, conn net.Conn) {
	InfoLogger.Println("TopicAddListenerConn called")
	// create subscribe request
	topicListenerConn := new(topicListenerConn)
	topicListenerConn.topicName = topicName
	topicListenerConn.listeningConn = conn
	// sending subscribe request
	node.topicListenerConnCh <- *topicListenerConn
	topicListenerConn = nil
}

// returns all topic names
func NodeListTopics(node Node) []string {
	InfoLogger.Println("NodeListTopics called")
	keys := make([]string, 0, len(node.topics))
	for k, _ := range node.topics {
		keys = append(keys, k)
	}
	if len(node.topics) > 0 {
		return keys
	} else if len(node.topics) == 0 {
		return []string{"none"}
	}
	return []string{"none"}
}

// creates pull request and sends it to the topic handler
// topic handler processes created request and sends result to given conn
func TopicPullData(node Node, conn net.Conn, topicName string, nmsg int) {
	InfoLogger.Println("TopicPullData called")
	// creating pull request
	topicPullReq := new(topicPullReq)
	topicPullReq.topicName = topicName
	topicPullReq.nmsg = nmsg
	topicPullReq.conn = conn
	// sending pull request to topic handler
	node.topicPullCh <- *topicPullReq
}

// creates push request and sends it to the topic handler
// topic handler processes created request and adds new topic msg to given topic name
func TopicPublishData(node Node, topicName string, tdata []byte) {
	InfoLogger.Println("TopicPublishData called")
	topicMsg := new(topicMsg)
	topicMsg.topicName = topicName
	topicMsg.msg = tdata
	node.topicPushCh <- *topicMsg
	topicMsg = nil
}

// creates topic create request and sends it to the topic handler
// topic handler processes created request and adds new topic to the topics map
func TopicCreate(node Node, topicName string) {
	InfoLogger.Println("TopicCreate called")
	topicName = rcf_util.ApplyNamingConv(topicName)

	node.topicCreateCh <- topicName
}

// creates action create request and sends it to the action handler
// action handler processes created request and adds new action with given action name
func ActionCreate(node Node, actionName string, actionFunc actionFn) {
	actionName = rcf_util.ApplyNamingConv(actionName)
	newAction := new(action)
	newAction.actionName = actionName
	newAction.actionFunction = actionFunc
	node.actionCreateCh <- *newAction
	newAction = nil
	InfoLogger.Println("ActionCreate called")
}

// creates actions execution request and sends it to the action handler
// action handler processes created request and executes the action
func ActionExec(node Node, actionName string, actionParams []byte) {
	InfoLogger.Println("ActionExec called")
	actionExec := new(actionExec)
	actionExec.actionName = actionName
	actionExec.params = actionParams
	node.actionExecCh <- *actionExec
	actionExec = nil
}

// creates service create request and sends it to the service handler
// service handler processes created request and adds new service with given service name
func ServiceCreate(node Node, serviceName string, serviceFunc serviceFn) {
	serviceName = rcf_util.ApplyNamingConv(serviceName)
	service := new(service)
	service.serviceName = serviceName
	service.serviceFunction = serviceFunc
	node.serviceCreateCh <- *service
	service = nil
	InfoLogger.Println("ServiceCreate called")
}

// creates service execution request and sends it to the service handler
// service handler processes created request and executes the service as well as returns the result to the given connection(client who called the service)
func ServiceExec(node Node, conn net.Conn, serviceName string, serviceParams []byte) {
	InfoLogger.Println("ServiceExec called")
	serviceExec := new(serviceExec)
	serviceExec.serviceName = serviceName
	serviceExec.serviceCallConn = conn
	serviceExec.params = serviceParams
	node.serviceExecCh <- *serviceExec
	serviceExec = nil
}
