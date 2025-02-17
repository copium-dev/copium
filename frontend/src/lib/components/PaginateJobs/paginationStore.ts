import { writable } from 'svelte/store';

export const paginationStore = writable({
    count: 0,
})