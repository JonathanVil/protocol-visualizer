class Actor {
    constructor(id) {
        this.id = id;
        this.nodeColor = '#1d4ed8';
    }


    // Called automatically when a message is delivered
    receive(msg) {
        if (msg.type === "PING") {

            // Send a message to one node
            send(this.id, msg.from, "PING");
        }
    }

    // Sends a PING message to another actor. Can be run manually from the actor.
    sendPing(receiverID) {
        send(this.id, receiverID, "PING");
    }
}

return Actor;
