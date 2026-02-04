// This function parses the user code and returns actors + messages
import ts from "typescript";
export function parseProtocolCode(codeString) {
    const actors = [];
    const messages = [];

    try {
        // Evaluate user code string into js code:
        const jsCode = ts.transpile(codeString);

        const result = new Function(jsCode)();

        const graphActors = [];
        const graphMessages = [];

        //Convert actors to graph nodes
        if (Array.isArray(result.actors)) {
            for (const actor of result.actors) {
                graphActors.push({
                    id: String(actor.id),
                    label: actor.name,
                })
            }
        }

        //Convert messages to graph edges
        if (Array.isArray(result.messages)) {
            for (const msg of result.messages) {
                graphMessages.push({
                    source: String(msg.source),
                    target: String(msg.target),
                    label: msg.label
                })
            }
        }

        return { actors: graphActors, messages: graphMessages };

        /* HARDCODED version that works
        const ActorClass = new Function(codeString)();


        actors.push({ id: 'A', label: 'Actor A' });
        actors.push({ id: 'B', label: 'Actor B' });
        actors.push({ id: 'C', label: 'Actor C' });
        actors.push({ id: 'D', label: 'Actor D' });


        messages.push({ source: 'A', target: 'B', label: 'msg1' });
        messages.push({ source: 'B', target: 'C', label: 'msg2' });
         */

    } catch (e) {
        console.error('Error parsing code:', e);
        return { actors: [], messages: [] };
    }
}
