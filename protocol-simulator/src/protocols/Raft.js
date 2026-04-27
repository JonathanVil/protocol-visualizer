// https://www.usenix.org/system/files/conference/atc14/atc14-paper-ongaro.pdf
class Actor {
    constructor(id) {
        this.id = id;
        this.nodeColor = 'blue';

        // STATE
        this.state = "FOLLOWER"
        this.currentTerm = 0;
        this.votedFor = null;
        this.log = [""];
        this.commitIndex = 0;
        this.lastApplied = 0;
        this.nextIndex = [];
        this.matchIndex = [];
        this.votes = 0;
        this.timeoutId = null
        this.stateCounter = 0;
        this.maxTransitTime = 40; // Used in setting timers. If the maximum transit time is higher than this, the system may fail
        this.startElectionTimeout();
    }

    receive(msg) {

        if (msg.type === 'VOTE_REQUEST') {
            /*
            msg.data: { term: number; candidateId: number; lastLogIndex: number; lastLogTerm: number; }
             */

            let vote = false;

            // a vote request in a new term, reset.
            if (msg.data.term > this.currentTerm) {
                this.votedFor = null;
                this.currentTerm = msg.data.term;
                this.becomeFollower()
            }

            // Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                vote = false;
            } else if (this.votedFor === null || this.votedFor === msg.data.candidateId) { // "Each server will vote for at most one candidate in a given term"

                this.startElectionTimeout();
                // Check if candidate's log is at least as up-to-date as receiver's log
                let upToDate = false;
                let lastLogTerm = this.log.length > 1 ? this.log[this.log.length - 1].term : 0;

                if (lastLogTerm < msg.data.lastLogTerm) {
                    // The last entry in the candidate's log has a higher term than the last entry in the receiver's log, see $5.4.1
                    upToDate = true;
                } else if (lastLogTerm === msg.data.lastLogTerm && this.log.length - 1 <= msg.data.lastLogIndex) {
                    // The last entry in both logs have the same term AND the candidate's log is at least as long as receiver's log
                    upToDate = true;
                }
                if (upToDate) {
                    this.votedFor = msg.data.candidateId;
                    this.currentTerm = msg.data.term;
                    vote = true;
                }
            }
            send(this.id, msg.from, 'VOTE_REPLY', { term: this.currentTerm, voteGranted: vote } );

        } else if (msg.type === 'VOTE_REPLY') {
            /*
            msg.data: { term: number; voteGranted: boolean; }
             */

            if (msg.data.term > this.currentTerm) {
                this.currentTerm = msg.data.term;
                this.becomeFollower()
            }

            if (msg.data.voteGranted) {
                this.votes++;
                // If votes received from majority of actors. It wins the election (a)
                if (this.votes >= Math.floor(getActors() / 2) + 1) {
                    this.becomeLeader();
                    // BEGIN SENDING HEARTBEATS
                }
            }


        } else if (msg.type === 'APPEND_ENTRIES_REQUEST') {
            /*
            msg.data: { term: number; prevLogIndex: number; prevLogTerm: number; entries: []; leaderCommit: number; }
             */

            // "While waiting for votes, a candidate may receive an AppendEntries RPC from another server claiming to be leader. If the leader’s term (included in its RPC) is at least
            // as large as the candidate’s current term, then the candidate recognizes the leader as legitimate and returns to follower state." (b)
            if (this.currentTerm <= msg.data.term) {
                this.currentTerm = msg.data.term;
                this.becomeFollower();
            }
            if (this.state !== "FOLLOWER") {return}


            this.startElectionTimeout()
            // 1. Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                let result = {
                    term: this.currentTerm,
                    success: false
                }
                send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', result);
                return;
            }

            //  2. Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm (§5.3)
            if (msg.data.prevLogIndex > 0 && (this.log.length <= msg.data.prevLogIndex || this.log[msg.data.prevLogIndex].term !== msg.data.prevLogTerm)) {
                let result = {
                    term: this.currentTerm,
                    success: false
                }
                send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', result);
                return;
            }

            // at this point we have accepted that this append request
            this.votedFor = msg.data.leaderId

            // 3. If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it (§5.3)
            if (msg.data.prevLogIndex === 0) { // if the whole log is bad
                this.log = [""];
            }
            if (this.log.length - 1 > msg.data.prevLogIndex) {
                this.log = this.log.slice(0, msg.data.prevLogIndex + 1);
            }

            // 4. Append any new entries not already in the log
            // [1, 2, 3] <- [2, 3, 4] = [1, 2, 3, 2, 3, 4]
            this.log = this.log.concat(msg.data.entries);
            let result = {
                term: this.currentTerm,
                success: true,
                newLogLength: this.log.length
            }

            if (msg.data.leaderCommit > this.commitIndex) {
                this.commitIndex = Math.min(msg.data.leaderCommit, this.log.length - 1);
            }
            if (this.commitIndex > this.lastApplied) {
                this.applyCommand()
            }

            send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', result);


        } else if (msg.type === "APPEND_ENTRIES_REPLY") {
            /*
            msg.data: { term: number; success: boolean; newLogLength: number?; }
             */



            // if the follower's term is higher than current leaders term
            if (msg.data.term > this.currentTerm) {
                this.currentTerm = msg.data.term;
                this.becomeFollower();
            }

            if (this.state !== "LEADER") return; // only leaders process append entries replies

            // Only true, if the followers log matches the leaders log.
            if (msg.data.success) {
                this.nextIndex[msg.from] = msg.data.newLogLength;
                this.matchIndex[msg.from] = msg.data.newLogLength - 1;

                // If there exists an N such that N > commitIndex ,a majority
                // of match Index[i] ≥= N,and log[N].term == currentTerm:
                // set commitIndex = N (§5.3,§5.4).
                // - Sort matchIndex to find the median matchIndex. Decrement if terms does not match && i > commitIndex
                let sorted = this.matchIndex.toSorted();
                for (let i = sorted[Math.floor(sorted.length / 2)]; i > this.commitIndex; i--) {
                    if (this.log[i].term === this.currentTerm) {
                        this.commitIndex = i;
                        break;
                    }
                }

                if (this.commitIndex > this.lastApplied) {
                    this.applyCommand()
                }
            } else {
                if (this.nextIndex[msg.from] > 1) {
                    this.nextIndex[msg.from] -= 1;
                }
                this.sendAppendEntries(msg.from)
            }

        } else if (msg.type === "REDIRECT") {
            this.receiveClientRequest(msg.data.command);
        }
    }

    receiveClientRequest(command) { // We have chosen not to represent the client node for this simulation. This is how we make requests instead
        // only leaders can receive client requests
        if (this.state !== "LEADER") {
            send(this.id, this.votedFor, "REDIRECT", command);
            return;
        }

        //1. Append command to the log
        this.log = [...this.log, { term: this.currentTerm, command: command }];

        for (let actorId = 0; actorId < getActors(); actorId++) {
            if (actorId !== this.id) {
                this.sendAppendEntries(actorId);
            }
        }
    }

    applyCommand() {

        this.lastApplied++;
        let command = this.log[this.lastApplied].command;
        if (command === "ADD") {
            this.stateCounter += 1;
        } else if (command === "SUBTRACT") {
            this.stateCounter -= 1;
        } else if (command !== "SHOW") {
            return
        }
        if (this.state === "LEADER") { // if there were a client node, this is where the leader would reply to their request
            console.log("Leader " + this.id + " applied command: " + command + ". Counter: " + this.stateCounter);
        }


    }

    becomeFollower() {
        this.state = "FOLLOWER";
        this.votes = 0;
        this.nodeColor = 'blue';
        this.startElectionTimeout()
    }

    becomeLeader() {
        this.state = "LEADER";
        this.votes = 0;
        this.nextIndex = Array(getActors()).fill(this.log.length);
        this.matchIndex = Array(getActors()).fill(0);
        deleteTimeout(this.timeoutId);
        this.timeoutId = null;
        this.nodeColor = 'red';
        this.broadcastAppendEntries()
    }

    becomeCandidate() {
        this.state = "CANDIDATE";
        this.currentTerm = this.currentTerm + 1;
        this.votedFor = this.id;
        this.votes = 1;
        this.nodeColor = 'orange';
        this.startElectionTimeout()
    }

    startElectionTimeout() {
        //the minimum time here should be heartbeat timer + max transit time.
        let randomTimeout = (10 + 3 * this.maxTransitTime + Math.floor(Math.random() * (2 * this.maxTransitTime))); // Random timeout between 80 and 160 ticks
        deleteTimeout(this.timeoutId);
        this.timeoutId = timeout(this, randomTimeout, this.requestVotes);
    }

    // send a VOTE_REQUEST to all other actors. Called upon election timeout.
    // "A candidate continues in this state until one of three things happens:
    // (a) it wins the election,
    // (b) another server establishes itself as leader, or
    // (c) a period of time goes by with no winner."
    requestVotes() {
        this.becomeCandidate();
        this.startElectionTimeout() // (c)
        let lastLogTerm = this.log.length > 1 ? this.log[this.log.length - 1].term : 0;
        for (let actorId = 0; actorId < getActors(); actorId++) {
            if (actorId !== this.id) {
                let voteRequestData = {
                    term: this.currentTerm,
                    candidateId: this.id,
                    lastLogIndex: this.log.length - 1,
                    lastLogTerm: lastLogTerm
                }
                send(this.id, actorId, "VOTE_REQUEST", voteRequestData);
            }
        }
    }

    // send a APPEND_ENTRIES_REQUEST from leader to followers. Also used for heartbeat.
    broadcastAppendEntries() {
        if (!this.state === "LEADER") return;

        for (let actorId = 0; actorId < getActors(); actorId++) {
            if (actorId !== this.id) {
                this.sendAppendEntries(actorId);
            }
        }

        this.timeoutId = timeout(this, (10 + 2 * this.maxTransitTime), this.broadcastAppendEntries); //should be set to at least 2x the maximum delivery time
    }

    sendAppendEntries(actorId) {
        if (actorId > this.nextIndex.length - 1) {
            this.nextIndex[actorId] = this.log.length;
        }

        let prevLogIndex = this.nextIndex[actorId] - 1;
        let prevLogTerm = 0
        let entries = this.log.slice(this.nextIndex[actorId]);
        if (prevLogIndex > 0)  {
            prevLogTerm = this.log[prevLogIndex].term;
        }

        let appendEntriesData = {
            term: this.currentTerm,
            leaderId: this.id,
            prevLogIndex: prevLogIndex,
            prevLogTerm: prevLogTerm,
            entries: entries,
            leaderCommit: this.commitIndex
        }
        send(this.id, actorId, "APPEND_ENTRIES_REQUEST", appendEntriesData);
    }

    revive() {
        this.votedFor = null
        this.becomeFollower()
    }
}

return Actor;
