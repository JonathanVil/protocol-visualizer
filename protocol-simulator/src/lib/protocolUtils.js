// This function parses the user code and returns actors + messages

/**
 * @param {string} codeString
 * @param {function} send
 * @param {function} getActors
 * @param {function} createQueue
 * @returns {ActorConstructor}
 */
export function parseProtocolCode(codeString, send, getActors, createQueue) {

    try {

        return new Function(
            "send",
            "getActors",
            "createQueue",
            codeString
        )(send, getActors, createQueue);

    } catch (e) {
        console.error('Error parsing code:', e);
        return null;
    }
}

export let transitTime = 10;
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
    return id_messages--;
}
