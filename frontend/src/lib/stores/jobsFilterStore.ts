import { writable } from 'svelte/store';

export const jobsFilterStore = writable({
    query: '',
    company: '',
    role: '',
    location: '',
    startDate: '',
    endDate: '',
    status: 'Status'
});