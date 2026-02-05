// This function parses the user code and returns actors + messages

/**
 * @param {string} codeString
 * @param {function} send
 * @returns {ActorConstructor}
 */
export function parseProtocolCode(codeString, send) {

    try {

        return new Function(
            "send",
            codeString
        )(send);

    } catch (e) {
        console.error('Error parsing code:', e);
        return null;
    }
}
