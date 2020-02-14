package test_client

import (
 rcf_cc_topic "robot-communication-framework/rcf_cc_topic"
)

func main() {

  rcf_cc_topic.Push_data("dsadsad", 200, "test")

  //time.Sleep(10)
}
