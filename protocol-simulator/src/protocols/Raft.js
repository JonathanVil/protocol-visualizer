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
        this.nextIndex = {};
        this.matchIndex = {};
        this.votes = 0;

        this.electionTimeout();
    }

    receive(msg) {

        if (msg.type === 'VOTE_REQUEST') {
            let vote = false;

            // a voterequest in a new term, reset.
            if (msg.data.term > this.currentTerm) {this.votedFor = null;}

            // Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                vote = false;
            } else if (this.votedFor === null || this.votedFor === msg.data.candidateId) { // "Each server will vote for at most one candidate in a given term"

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
            send(this.id, msg.from, 'VOTE_REPLY', [this.currentTerm, vote] );

        } else if (msg.type === 'VOTE_REPLY') {
            this.votes++;
            // If votes received from majority of actors. It wins the election (a)
            if (this.votes > Math.floor(getActors() + 1 / 2)) {
                this.state = "LEADER";
                this.votes = 0;
                // BEGIN SENDING HEARTBEATS
            }
        } else if (msg.type === 'APPEND_ENTRIES_REQUEST') {

            // "While waiting for votes, a candidate may receive an AppendEntries RPC from another server claiming to be leader. If the leader’s term (included in its RPC) is at least
            // as large as the candidate’s current term, then the candidate recognizes the leader as legitimate and returns to follower state." (b)
            if (this.state === "CANDIDATE" && this.currentTerm <= msg.data.term) {
                this.state = "FOLLOWER";
                this.votes = 0;
            }

            // 1. Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', false);
                return;
            }

            //  2. Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm (§5.3)
            if (this.log.length <= msg.data.prevLogIndex || this.log[msg.data.prevLogIndex].term !== msg.data.prevLogTerm) {
                send(this.id, msg.from, 'APPEND_ENTRIES_REPLY', false);
                return;
            }

            // 3. If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it (§5.3)
            


        }


    }

    electionTimeout() {
        let randomTimeout = (150 + Math.floor(Math.random() * 150)) / 3; // Random timeout between 50 and 100 ticks
        timeout(this, randomTimeout, this.requestVote);
    }

    // send a VOTEREQUEST to all other actors. Called upon election timeout.
    // "A candidate continues in this state until one of three things happens:
    // (a) it wins the election,
    // (b) another server establishes itself as leader, or
    // (c) a period of time goes by with no winner."
    requestVote() {
        if (this.votedFor !== null && this.votedFor !== this.id) return; // already voted for someone else, do not start election.
        this.currentTerm = this.currentTerm + 1;
        this.state = "CANDIDATE";
        this.votedFor = this.id
        this.electionTimeout() // (c)
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

    // send a APPENDENTRIESEQUEST from leader to followers. Also used for heartbeat.
    sendAppendEntries() {
        let appendEntryData = {
            term: this.currentTerm,
            leaderId: this.id,
            prevLogIndex: this.prevLogIndex,
            prevLogTerm: this.prevLogTerm,
            entries: this.log
        }
        for (let actorId = 0; actorId < getActors(); actorId++) {
            if (actorId !== this.id) {
                send(this.id, actorId, "APPEND_ENTRIES_REQUEST", appendEntryData);
            }
        }
    }
}

return Actor;
