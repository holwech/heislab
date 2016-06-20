# Heislab - TTK4145 Sanntidsprogrammering - Real-Time Programming

## Modules
This project consists of 7 parts.

### Slave
Slave initializes all other modules. It sets the master-module locally or joins any master-module already online on the local network.

### Master
Master is the "brain" that controls everything. Only one brain is active at any time, but every running program has a idle master that is in sync with the active master. If a computer loses connection, the idle master will become active and pick up from where on what the previous master was doing.

### Network
Works as an interface between master/slave and the communication-module. Network sorts messages between slave and master. It also does some simple message convertion.

### Orders
Orders calculates orders based on a cost function. This module is used by master to decide where each elevator should go.

### Driver
Driver works as an interface between hardware and software. Button clicks and other input can be accessed from the driver module.

### Communication
Communication is a simple UDP-module that receives messages from any network access. All messages are sent over UDP-broadcast. Includes a simple receive-confirmation functionality that passed a error message if the message is not received.

### cl (command-list)
A list of commands that slave and master use to communicate with each other.
