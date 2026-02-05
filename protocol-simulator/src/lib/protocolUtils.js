// This function parses the user code and returns actors + messages

/** @returns {ActorConstructor}*/
export function parseProtocolCode(codeString) {

    try {

        return new Function(codeString)();

    } catch (e) {
        console.error('Error parsing code:', e);
        return null;
    }
}
