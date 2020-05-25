# Robot Communication Framework

The RCF is a framework for data distribution, which is the most essential part of an autonomous platform. It is very similar to [ROS](https://www.ros.org/) but without packages and the C/C++ complexity overhead while still maintaining speed and **safe** thread management standards, thanks to the [go](https://golang.org/) lang.

# Philosophy

Communication and concurrency is one of the prominent challenges when developing a robot platform. Since every robot can have multiple kinds of sensors and control interfaces it has to interact with, the communication between those elements is an absolute indispensable part of the foundation of every robot.
Considering such rather complex requirements, the probability of failure increases due to the increasing complexity. So the goal of this project is not just to keep the complexity low but to increase the reliability and maintainability of the framework. This is achieved by using only core feature and shrinking the dependencies to an absolute minimum. Another aspect that needs to be considered is the chosen programming language and its reliability/ dependency. This projects uses go with absolutely no dependencies, using only core features for concurrency and networking. Another great thing about go is that it does not require a virtual runtime of any kind and integrates more or less bar metal in the system(as for example C/C++). A further great feature is its concurrency support, which is achieved by eliminating parallelization and replacing it with a channel interface, with which data can be shared between functions. That is done by guaranteeing that the data is handled, if done not so, the program will deadlock to prohibit any memory parallelization issues. 

# Concept

The primary communication interface is a node, in contrast to ROS, or various other robot platforms. A node resembles the platform for services, actions and topics which can be accessed by the rcf clients.
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
