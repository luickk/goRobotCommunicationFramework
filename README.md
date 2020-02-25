# Robot Communication Framework

The RCF is a framework for data distribution, which is the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages and the C/C++ complexity overhead while still maintaining speed and **safe** thread/ lang. standards, thanks to the [go](https://golang.org/) lang.

# Goroutine Memory Synchronisation

Since sharing memory is a very complicated and difficult thing to get right without overcompicating things, Go channels are used, according to golang's motto "Share memory by communicating, don't communicate by sharing memory"!

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

The primary communication interface is a node, in contrast to ROS, or various other robot platforms, the node is only a node objects and does not contain any code. The code and its functionality is contained inside basic go routines.
A node resembles the platform for services and topics.

Every node has a port number which also resembles the node ID, but no actual name, so every node is represented through a number. This is a major part of the concept to reduce complexity since there is no need for inter node communication to resolve addresses.

## Topics

Every topic represents a communication channel/ queue from which data can be popped or pushed onto.
A topic can be identified via its name and the node(node ID) which it is hosted by.

## Command & Control data

C&C topics and services are shared via TCP servers/ endpoints which are resembled by nodes. Via a topic data such as sensor, command&control data can be shared and accessed by everybody who can connect to a node.

## Error tolerant data

Video, image or audio data that's error tolerant and requires high bandwidth can be shared via streamed topics, which are very similar to normal topics, with the only difference that the data is shared via UDP instead of TCP.    
