/** @typedef {new (id: number) => Actor} ActorConstructor */

/** @typedef {{ id: number, source: number, destination: number, type: string, transitSteps: number, elapsedSteps: number}} Message */
// TODO: allow messages to carry a value

/**
 * @typedef {{
 *   id: number,
 *   receive(message: { type: string, from: number }): void
 * }} Actor
 */

export {};