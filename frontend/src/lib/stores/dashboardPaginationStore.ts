import { writable } from 'svelte/store';

export const dashboardPaginationStore = writable({
    perPage: 0,
    count: 0,
})