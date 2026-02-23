
// This function parses the user code and returns actors + messages

/**
 * @param {string} codeString
 * @param {function} send
 * @param {function} getActors
 * @param {function} createQueue
 * @param {function} timeout
 * @returns {ActorConstructor|null}
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

let transitTimeUpperBound = 20;
let transitTimeLowerBound = 20;

/**
 * @return {number} The transit time in ticks
 */
export function getTransitTime() {
    return Math.floor(Math.random() * (transitTimeUpperBound - transitTimeLowerBound + 1)) + transitTimeLowerBound;
}

export function setTransitbounds(highBound) {
    transitTimeUpperBound = highBound;
}


let tickSize = 100;
export function getTickSize() {
    return tickSize;
}

export function getTickSpeed() {
    return 1000 / tickSize;
}

export function setTickSpeed(value) {
    tickSize = 1000 / value;
}


// id's for messages
/** @type {number} */
let id_messages = -1;

/**
 * @return {number} The next message id
 */
export function getNextMessageId() {
    if (id_messages < -1000) {id_messages = -1}
    return id_messages--;
}
