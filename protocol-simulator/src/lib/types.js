/** @typedef {new (id: number) => Actor} ActorConstructor */

/** @typedef {{ id: number, source: number, destination: number, type: string, transitTicks: number, elapsedTicks: number, data?: any}} Message */

/** @typedef {{ actorId: number, ticks: number, reaction: function}} TimeoutEntry */


/**
 * @typedef {{
 *   id: number,
 *   receive(message: { type: string, from: number }): void,
 *   nodeColor: string
 * }} Actor
 */

export {};