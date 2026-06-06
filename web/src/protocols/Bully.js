class Actor {
    constructor(id) {
        this.id = id;
        this.status = "";
        this.leader = null;
        this.leaderIsAlive = false;
        this.coordinationSent = false;
        this.seqNum = 0;
        this.expectedSeqNums = []
        this.maxExpectedTransit = 30 // If max transit time is higher than this number, Bully may fail.
    }

    receive(msg) {

        //Check if the seq number of the message is as expected
        //If not we drop (ensures messages are retrieved in same order as they were send)
        if ( msg.data <= this.expectedSeqNums[msg.from]) {
            console.log("Messages had a lower sequence number... ignoring");
            return;
        }
        this.expectedSeqNums[msg.from] = msg.data

        if (msg.type === "PING") { //an empty message, serves as a heartbeat
            send(this.id, msg.from, "PONG", this.incSeq())

        }
        if (msg.type === "PONG") { // response to the heartbeat
            this.leaderIsAlive = true;
        }

        if (msg.type === "ELECTION") { // sent when the leader is discovered dead
            this.leaderIsAlive = false
            send(this.id, msg.from, "REPLY", this.incSeq())
            this.startElection()
        }

        if (msg.type === "REPLY") { // reply to an election message
            this.leaderIsAlive = true;
            if (this.leader === null || this.leader < msg.from) {
                this.leader = msg.from
                this.nodeColor = '#1d4ed8';
            }
            this.status = "Reorganization"
        }
        if (msg.type === "COORDINATE") {
            this.leader = msg.from
            this.nodeColor = '#1d4ed8';
            this.status = "Normal"
        }

    }

    startElection(){
        if (!this.leaderIsAlive) {
            this.leader = this.id
            this.status = "Election"
            this.nodeColor = '#82242fff';

            for (let i = this.id + 1; i < getActors(); i++) {
                send(this.id, i, "ELECTION", this.incSeq());
            }

            //After x ticks, it tries to coordinate as leader.
            timeout(this, 2 * this.maxExpectedTransit, this.coordinate);

        }
    }
    coordinate(){
        //Only coordinate if this node is still leader.
        if (this.leader === this.id && this.coordinationSent === false){
            for (let i = 0; i < getActors(); i++) {
                if (i != this.id) {
                    send(this.id, i, "COORDINATE", this.incSeq());
                }

            }
            this.coordinationSent = true;
            timeout(this, 2 * this.maxExpectedTransit, this.finishCoordinate);
        }

    }
    finishCoordinate(){
        this.coordinationSent = false;
    }

    incSeq() {
        this.seqNum = this.seqNum + 1;
        return this.seqNum;
    }

    sendPing() {
        //If leader is alive
        if (this.leader != null) {
            this.leaderIsAlive = false; //set to false in case leader is dead, if leader is alive it will be set to true again
            send(this.id, this.leader, "PING", this.incSeq())
            timeout(this, 2 * this.maxExpectedTransit, this.startElection)
        } else {
            this.startElection()
        }
    }
}

return Actor;
