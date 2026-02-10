
// This function parses the user code and returns actors + messages

/**
 * @param {string} codeString
 * @param {function} send
 * @param {function} getActors
 * @param {function} createQueue
 * @param {function} timeout
 * @returns {ActorConstructor}
 */
export function parseProtocolCode(codeString, send, getActors, createQueue, timeout) {

    try {

        return new Function(
            "send",
            "getActors",
            "createQueue",
            "timeout",
            codeString
        )(send, getActors, createQueue, timeout);

    } catch (e) {
        console.error('Error parsing code:', e);
        return null;
    }
}

let transitTimeUpperBound = 12;
let transitTimeLowerBound = 8;


export function getTransitTime() {
    return Math.floor(Math.random() * (transitTimeUpperBound - transitTimeLowerBound + 1)) + transitTimeLowerBound;
}

export function setTransitbounds(highBound, lowBound) {
    transitTimeUpperBound = highBound;
    transitTimeLowerBound = lowBound;
}


let stepSize = 100;
export function getStepSize() {
    return stepSize;
}

export function setStepSize(value) {
    stepSize = value;
}

// id's for messages
/** @type {number} */
let id_messages = -1;

export function getNextMessageId() {
    if (id_messages < -1000) {id_messages = -1}
    return id_messages--;
}
