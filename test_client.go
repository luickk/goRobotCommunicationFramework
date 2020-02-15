package main

import (
 rcf_cc_topic "robot-communication-framework/rcf_cc_topic"
)

func main() {

  // rcf_cc_topic.Push_data("dsadsad", 200, "test")
  rcf_cc_topic.Pop_data(2, 200, "test")

}
