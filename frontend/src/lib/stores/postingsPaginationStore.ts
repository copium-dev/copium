import { writable } from 'svelte/store';

export const postingsPaginationStore = writable({
    perPage: 0,
    count: 0,
})