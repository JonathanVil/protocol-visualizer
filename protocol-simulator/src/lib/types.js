/** @typedef {new (id: number) => Actor} ActorConstructor */

/** @typedef {{ id: number, source: number, destination: number, type: string, sentTick: number, arrivalTick: number, data?: any}} Message */

/** @typedef {{ actorId: number, ticks: number, reaction: function}} TimeoutEntry */


/**
 * @typedef {{
 *   id: number,
 *   receive(message: { type: string, from: number }): void,
 *   nodeColor: string
 *   alive: boolean
 *   protocolName: string|null
 * }} Actor
 */

/**
 * @typedef {{
 *   id: string,
 *   name: string,
 *   model: import('monaco-editor').editor.ITextModel;
 *   viewState?: import('monaco-editor').editor.ICodeEditorViewState | null;
 * }} EditorTab
 */

export {};