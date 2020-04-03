# Robot Communication Framework

The RCF is a framework for data distribution, which is the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages and the C/C++ complexity overhead while still maintaining speed and **safe** thread/ lang. standards, thanks to the [go](https://golang.org/) lang.

Examples can be found in ./examples.

# Concept

The primary communication interface is a node, in contrast to ROS, or various other robot platforms, the node is only a object instance and does not contain any code. A node resembles the platform for services, actions and topics.


Every node has a port number which also resembles the node ID, but no actual name, so every node is represented through a number. This is a major part of the concept to reduce complexity since there is no need for internal node communication to resolve addresses.
# Installation

Installation via. command line: <br>

`cd Users/<username>/go/src/ ` <br>
`git clone https://github.com/cy8berpunk/rcf` <br>

# Go-Routine Memory synchronization

Since sharing memory is a very complicated and difficult thing to get right without overcomplicating things, Go channels are used, according to Golang's motto "Share memory by communicating, don't communicate by sharing memory"!

## Topics

Every topic represents a communication channel, from which data can be pulled from or pushed onto or to which can be listened.
The topic communication is split up into msg's, every msg represents a byte array pushed to the topic. Every node has a topic msg capacity, so only the last x msg's are stored. There are no variable assignments, a msg can represent a single value or anything else encoded into a byte array. If a topic msg structure is needed, the `glob` methods serialize a string map and use the serialized maps as msgs and as such enable a structured and more generic way to use topics.
A topic is meant to share command & control or sensor data, hence data that needs to be accurate and which does not require high bandwith, since a topic rely's on tcp sockets to communicate.
A topic can be identified via its name and the node(node ID) which it is hosted by.

## Actions

An action is a function that can be executed, with parameters, by nodes or node clients. Since they are not meant to do calculations but to provide node side functionality for the clients, they can only be called without a return value.
Has to be declared on the node side.

## Services

A service is a function that can be executed with parameters and process for a certain amount of time, to finally return a result in form of a byte array.
Has to be declared on the node side.

## Protocols

#### Delimiter

Single commands are separated by a "\n". Msgs, encoded in commands(protocol based) are also separated by a "\n" (for example the glob maps).

#### Node

`><type>-<name>-<operation>-<paypload byte slice>`

#### Client Read Protocol
`><type>-<name>-<len(msgs)>-<paypload(msgs) byte slice>`
