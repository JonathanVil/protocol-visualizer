// https://www.usenix.org/system/files/conference/atc14/atc14-paper-ongaro.pdf
class Actor {
    constructor(id) {
        this.id = id;

        // STATE
        this.currentTerm = 0;
        this.votedFor = null;
        this.log = [];
        this.commitIndex = 0;
        this.lastApplied = 0;
        this.nextIndex = {};
        this.matchIndex = {};



    }

    receive(msg) {

        if (msg.type === 'VOTEREQUEST') {
            let vote = false;
            // Reply false if term < currentTerm (§5.1)
            if (msg.data.term < this.currentTerm) {
                vote = false;
            } else if (this.votedFor === null || this.votedFor === msg.data.candidateId) {

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
            send(this.id, msg.from, [this.currentTerm, vote], 'VOTEREPLY');
        }

    }

    //send a VOTEREQUEST to all other actors. Called upon election timeout.
    requestVote() {
        this.votedFor = this.id
        for (const actor in actors) {
            if (actor.id !== this.id) {
                let lastLogTerm = this.log.length > 0 ? this.log[this.commitIndex - 1].term : 0;
                let voteRequestData = {
                    term: this.currentTerm,
                    candidateId: this.id,
                    lastLogIndex: this.commitIndex,
                    lastLogTerm: lastLogTerm
                }
                send(this.id, actor.id, voteRequestData, 'VOTEREQUEST');
            }
        }
    }
}

return Actor;
