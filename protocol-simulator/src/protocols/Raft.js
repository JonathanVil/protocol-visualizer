// https://www.usenix.org/system/files/conference/atc14/atc14-paper-ongaro.pdf
class Actor {
    constructor(id) {
        this.id = id;

        // STATE
        this.state = "FOLLOWER"
        this.currentTerm = 0;
        this.votedFor = null;
        this.log = [];
        this.commitIndex = 0;
        this.lastApplied = 0;
        this.nextIndex = [];
        this.matchIndex = [];
        this.votes = 0;
        this.timeoutId = null
        this.startElectionTimeout();

    } // we are still figuring out how to make it work while the log is empty. Last changes were 

    receive(msg) {

        if (msg.type === 'VOTE_REQUEST') {
            /*
            msg.data: { term: number; candidateId: number; lastLogIndex: number; lastLogTerm: number; }
             */

            let vote = false;

            // a vote request in a new term, reset.
            if (msg.data.term > this.currentTerm) {
                this.votedFor = null;
            }

            // Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                vote = false;
            } else if (this.votedFor === null || this.votedFor === msg.data.candidateId) { // "Each server will vote for at most one candidate in a given term"

                this.startElectionTimeout();
                // Check if candidate's log is at least as up-to-date as receiver's log
                let upToDate = false;
                let lastLogTerm = this.log.length > 0 ? this.log[this.commitIndex - 1].term : 0;

                if (lastLogTerm < msg.data.lastLogTerm) {
                    // The last entry in the candidate's log has a higher term than the last entry in the receiver's log, see $5.4.1
                    upToDate = true;
                } else if (lastLogTerm === msg.data.lastLogTerm && this.commitIndex <= msg.data.lastLogIndex) {
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

            if (msg.data.voteGranted) {
                this.votes++;
                // If votes received from majority of actors. It wins the election (a)
                if (this.votes >= Math.floor(getActors() / 2) + 1) {
                    this.becomeLeader();
                    // BEGIN SENDING HEARTBEATS
                }
            } else if (msg.data.term > this.currentTerm) {
                this.currentTerm = msg.data.term;
            }


        } else if (msg.type === 'APPEND_ENTRIES_REQUEST') {
            /*
            msg.data: { term: number; prevLogIndex: number; prevLogTerm: number; entries: []; leaderCommit: number; }
             */

            this.startElectionTimeout();

            // "While waiting for votes, a candidate may receive an AppendEntries RPC from another server claiming to be leader. If the leader’s term (included in its RPC) is at least
            // as large as the candidate’s current term, then the candidate recognizes the leader as legitimate and returns to follower state." (b)
            if (this.state === "CANDIDATE" && this.currentTerm <= msg.data.term) {
                this.becomeFollower();
            }

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
            if (msg.data.prevLogIndex !== null && (this.log.length <= msg.data.prevLogIndex || this.log[msg.data.prevLogIndex].term !== msg.data.prevLogTerm)) {
                let result = {
                    term: this.currentTerm,
                    success: false
                }
                send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', result);
                return;
            }

            // 3. If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it (§5.3)
            if (msg.data.prevLogIndex === null) { // if the whole log is bad
                this.log = []
            }
            if (this.log.length > msg.data.prevLogIndex + 1 && this.log[msg.data.prevLogIndex + 1].term !== msg.data.term) {
                this.log = this.log.slice(0, msg.data.prevLogIndex + 1);
            }

            // 4. Append any new entries not already in the log
            // [1, 2, 3] <- [2, 3, 4] = [1, 2, 3, 2, 3, 4]
            this.log = this.log.concat(msg.data.entries);

            let result = {
                term: this.currentTerm,
                success: true
            }

            if (msg.data.leaderCommit > this.commitIndex) {
                this.commitIndex = Math.min(msg.data.leaderCommit, this.log.length);
                if (this.commitIndex > this.lastApplied) {
                    this.lastApplied++;
                    // TODO: apply log[this.lastApplied] to state machine
                }
            }

            send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', result);


        } else if (msg.type === "APPEND_ENTRIES_REPLY") {
            /*
            msg.data: { term: number; success: boolean; }
             */

            if (this.state !== "LEADER") return; // only leaders process append entries replies

            // if the follower's term is higher than current leaders term
            if (msg.data.term > this.currentTerm) {
                this.becomeFollower();
            }

            // Only true, if the followers log matches the leaders log.
            if (msg.data.success) {
                this.nextIndex[msg.from] = this.log.length;
                this.matchIndex[msg.from] = this.log.length - 1;

                // If there exists an N such that N > commitIndex ,a majority
                // of match Index[i] ≥= N,and log[N].term == currentTerm:
                // set commitIndex = N (§5.3,§5.4).
                // - Sort matchIndex to find the median matchIndex. Decrement if terms does not match && i > commitIndex
                let sorted = this.matchIndex.toSorted();
                for (let i = sorted[Math.floor(getActors() / 2)]; i > this.commitIndex; i--) {
                    if (this.log[i].term === this.currentTerm) {
                        this.commitIndex = i;
                        break;
                    }
                }

                if (this.commitIndex > this.lastApplied) {
                    this.lastApplied++;
                    // TODO: apply log[this.lastApplied] to state machine
                }
            } else {
                if (this.nextIndex[msg.from] > 0) {
                    this.nextIndex[msg.from] -= 1;
                }
                this.sendAppendEntries(msg.from)
            }

        }
    }

    recieveClientRequest(data) {
        
    }

    becomeFollower() {
        this.state = "FOLLOWER";
        this.votes = 0;
        this.nodeColor = 'blue';
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
        this.votedFor = this.id
        this.votes = 1;
        this.nodeColor = 'orange';
    }

    startElectionTimeout() {
        let randomTimeout = (100 + Math.floor(Math.random() * 100)); // Random timeout between 50 and 100 ticks
        deleteTimeout(this.timeoutId);
        this.timeoutId = timeout(this, randomTimeout, this.requestVotes);
    }

    // send a VOTE_REQUEST to all other actors. Called upon election timeout.
    // "A candidate continues in this state until one of three things happens:
    // (a) it wins the election,
    // (b) another server establishes itself as leader, or
    // (c) a period of time goes by with no winner."
    requestVotes() {
        if (this.votedFor !== null && this.votedFor !== this.id) return; // already voted for someone else, do not start election.
        this.becomeCandidate();
        this.startElectionTimeout() // (c)
        let lastLogTerm = this.log.length > 0 ? this.log[this.commitIndex - 1].term : 0;
        for (let actorId = 0; actorId < getActors(); actorId++) {
            if (actorId !== this.id) {
                let voteRequestData = {
                    term: this.currentTerm,
                    candidateId: this.id,
                    lastLogIndex: this.commitIndex,
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

        this.timeoutId = timeout(this, 50, this.broadcastAppendEntries); //might want to adjust all the timeouts to make it seem less busy
    }

    sendAppendEntries(actorId) {
        let prevLogIndex = null
        let prevLogTerm = null
        let entries = []
        if (this.nextIndex[actorId] > 0)  {
            prevLogIndex = this.nextIndex[actorId] - 1;
            prevLogTerm = this.log[prevLogIndex].term;
            entries = this.log.slice(this.nextIndex[actorId]);
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
}

return Actor;
