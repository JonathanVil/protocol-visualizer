// This function parses the user code and returns actors + messages
export function parseProtocolCode(codeString) {
    const actors = [];
    const messages = [];

    try {
        // Sandbox: run user code
       // const ActorClass = new Function(codeString)();


        actors.push({ id: 'A', label: 'Actor A' });
        actors.push({ id: 'B', label: 'Actor B' });
        actors.push({ id: 'C', label: 'Actor C' });
        actors.push({ id: 'D', label: 'Actor D' });


        messages.push({ source: 'A', target: 'B', label: 'msg1' });
        messages.push({ source: 'B', target: 'C', label: 'msg2' });

    } catch (e) {
        console.error('Error parsing code:', e);
    }

    return { actors, messages };
}
