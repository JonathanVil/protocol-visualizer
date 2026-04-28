import { writable } from 'svelte/store';
import { Queue } from '$lib/datastructures/Queue.js';

export const timeoutsStore = writable(new Queue());