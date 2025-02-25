import { writable } from 'svelte/store';

export const postingsPaginationStore = writable({
    count: 0,
})