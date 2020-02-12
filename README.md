# Robot Communication Framework

The RCF is a framework for data distribution, which the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages, the C/C++ complexity overhead while still maintaining speed and **safe** thread/ lang. standards, thanks the to [go](https://golang.org/) lang.

# Installation

Installation via. command line: <br>

`go get https://github.com/cy8berpunk/robot-communication-framework` <br>

installation from code

`
import (
    "fmt"
    "github.com/cy8berpunk/robot-communication-framework"
)
`

# Basic concept

The primary communication interface is a node, in contrast to ROS, or various other robot platforms, the node is only a node objects and does not contain any code. The code and its functionality is contained inside basic go instances or threads.
A node is always required if topics are created or services are provided.

Every node has a port number, but no actual name, so every node is represented through a number, this is a major part of the concept to reduce complexity.

## Command & Control data

C&C topics and services are shared via TCP servers/ endpoints which are resembled by nodes. Via a topic data such as sensor, command&control data can be shared and accessed by everybody who can connect to a node.

## Error tolerant data

Video, image or audio data that's error tolerant and requires high bandwidth can be shared via streamed topics, which are very similar to normal topics, with the only difference that the data is shared via UDP instead of TCP.    
