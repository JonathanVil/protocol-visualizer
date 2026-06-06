# System Basics
Welcome!
This is a tool for visualizing, simulating, debugging and experimenting with distributed protocols.
The tool allows you to define protocols using plain JavaScript, and to put actors running these protocols in to a 
simulated network to see how they interact.

## Actors
Actors are an abstraction of a machine running a protocol. Certain protocols can be selected in the code editor, or you can write your own!
Actors interact with other actors by sending and receiving messages across connections.
You can see the state and methods in an actor by pressing the icons at the top of the box next to the actor.
Through this box, you can also modify the state by pressing the small pencil next to a field, or run a method by pressing the small arrow.

## Ticks
The simulator is built on a "tick" system, like you might see in a video game. 
Ticks serve as the smallest unit of time in the simulation. 

## Messages
Every tick each message currently in transit moves one tick closer to its recipient. 
The amount of ticks in transit depends on the transit time, which may be modified through the UI. 
Once a message reaches the end, it is delivered and the recipient will handle it according to the protocol.

## Manually sending messages
You can manually send a message from one actor to another using the panel in the bottom left.
You will need to specify a sender id, target id, type, and, optionally, data.

## Controling the simulation
In order to start the simulation you can use the controls in the bottom right. 
Some basic settings are also available in the top right.

## The log
The bottom tab on the top left is the log. Here you can see a list of all events that have occured on a given tick.
You can also use the log to rewind to a previous state.

# Implementation Basics
The first step in simulating a protocol is the implementation of the actors.
A correct implementation must have several qualities.
It must provide a class specification called `Actor`, and a constructor for this class. It must also implement a `receive(msg)` method.
Lastly, it must return the class at the end of the program.
Provided below is a bare minimum implementation:
```js
class Actor {
    constructor(id) {
        this.id = id;
    }
    receive(msg) {
        if (msg.type === x) {
            //do something
        }
    }
}
return Actor
```

You can access the code editor through the middle tab on the top left.

## Fields
An actor must have an `id` field. This allows other actors to send messages to it. 
As long as you add the field, the simulation will automatically assign a unique id.
Other fields may be added in the constructor.
Setting the `nodeColor` field like: `this.nodeColor = '#1d4ed8'` 
allows you to customize the color of a node that is used in the visualization.

## Receive
The receive method should match on every type of message you expect an actor to receive.

For inspiration on how to implement a protocol yourself, 
you can load in some pre-implemented protocols by using the dropdown menu at the top of the screen.


# Built-in functions
In order to facilitate implementing a wide variety of protocols, the tool comes with a number of built-in functions.
## `send(from, to, type, ?data)`
This function allows sending messages between nodes.

It takes four arguments:
- `from` should always be the sender's ID, which can be accessed with `this.id`. This enables the recipient to reply if necessary.
- `to` is the recipient's ID.
- `type` is a string that specifies the kind of message being sent. In practice, it is used for matching messages in the `receive` method.
- `data` is an optional field that allows you to send an object containing additional information, such as timestamps.

## `getActors()`
returns the number of actors currently in the systems

## `createQueue(id, name)`
returns an empty linked list queue.

The queue has some methods for you to use:
- `queue.push(element)` allows you to push one element to the back of the queue.
- `queue.pop()` allows you tot pop one element from the front of the queue.
- `queue.length` gives the amount of elements in the queue.

## `timeout(actor, ticks, reaction)`
This function allows calling a method after a set amount of ticks.

It takes three arguments:
- `actor` should always be `this`. This is important to ensure the method can access state.
- `ticks` is the number of ticks the timer waits before executing the reaction.
- `reaction` is a reference to the method to be called when the timer finishes. This method must take no arguments.

The timeout function returns the id of the created timeout. 
The method 'deleteTimeout(id)' can be used to cancel the timer.
