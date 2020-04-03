
## Function Documentation

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
