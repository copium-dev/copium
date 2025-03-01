import { writable } from 'svelte/store';

export const postingsFilterStore = writable({
    query: '',
    company: '',
    role: '',
    location: '',
    startDate: '',
    endDate: '',
});