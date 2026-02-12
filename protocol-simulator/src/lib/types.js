/** @typedef {new (id: number) => Actor} ActorConstructor */

/** @typedef {{ id: number, source: number, destination: number, type: string, transitSteps: number, elapsedSteps: number, data?: any}} Message */
// TODO: allow messages to carry a value

/**
 * @typedef {{
 *   id: number,
 *   receive(message: { type: string, from: number }): void,
 *   nodeColor: string
 * }} Actor
 */

export {};