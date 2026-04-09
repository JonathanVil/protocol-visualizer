class Actor {
    constructor(id) {
        this.id = id;
        this.nodeColor = '#1d4ed8';

        this.ourSequenceNumber = 0;
        this.highestSequenceNumber = 0;

        this.requestingCS = false;
        this.accessingCS = false;
        this.messageQueue = createQueue();

        this.outstandingReplyCount = 0;
    }

    receive(msg) {
        if (msg.type === "REQUEST") {
            let k = msg.data;
            this.highestSequenceNumber = Math.max(this.highestSequenceNumber, k);
            if (this.requestingCS && ((k > this.ourSequenceNumber) || (k === this.ourSequenceNumber && this.id < msg.from))) { //crazy conditional
                this.messageQueue.push(msg.from) //DENIED
            } else {
                send(this.id, msg.from, "REPLY", this.ourSequenceNumber) //approved :)
            }
        }
        if (msg.type === "REPLY") {
            this.outstandingReplyCount--;
            if (this.outstandingReplyCount === 0){ //if everyone has approved, access the CS
                this.accessCS()
            }
        }
    }

    requestCS() {
        this.requestingCS = true;
        this.ourSequenceNumber = this.highestSequenceNumber + 1;
        this.outstandingReplyCount = getActors() - 1;
        for (let i = 0; i < getActors(); i++) {
            if (i === this.id) {
                continue
            }
            send(this.id, i, "REQUEST", this.ourSequenceNumber);
        }
    }

    accessCS() {
        console.log("Accessing CS");
        this.requestingCS = false;
        this.accessingCS = true;
        timeout(this, 20, this.exitCS)
    }

    exitCS() {
        console.log("Exiting CS");
        this.accessingCS = false;
        while (this.messageQueue.length > 0) {
            let id = this.messageQueue.pop()
            send(this.id, id, "REPLY");
        }
    }
}

return Actor;
