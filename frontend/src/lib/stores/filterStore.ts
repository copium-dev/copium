import { writable } from 'svelte/store';

export const filterStore = writable({
    query: '',
    company: '',
    role: '',
    location: '',
    startDate: '',
    endDate: '',
    status: 'Status'
});