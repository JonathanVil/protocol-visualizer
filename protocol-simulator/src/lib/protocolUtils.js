// This function parses the user code and returns actors + messages
import ts from "typescript";
export function parseProtocolCode(codeString) {
    const actors = [];

    try {
        // Evaluate user code string into js code:
        const jsCode = ts.transpile(codeString);

        const result = new Function(jsCode)();

        const graphActors = [];

        //Convert actors to graph nodes
        if (Array.isArray(result.actors)) {
            for (const actor of result.actors) {
                graphActors.push({
                    id: String(actor.id),
                    label: actor.name,
                })
            }
        }

        return { actors: graphActors};

    } catch (e) {
        console.error('Error parsing code:', e);
        return { actors: []};
    }
}
