/** @typedef {new (id: number) => Actor} ActorConstructor */

/** @typedef {{ id: number, source: number, destination: number, type: string, sentTick: number, arrivalTick: number, data?: any}} Message */

/**
 * @typedef {{
 *   id: number,
 *   receive(message: { type: string, from: number }): void,
 *   nodeColor: string
 * }} Actor
 */

export {};