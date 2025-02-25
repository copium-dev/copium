import { writable } from 'svelte/store';

export const dashboardPaginationStore = writable({
    count: 0,
})