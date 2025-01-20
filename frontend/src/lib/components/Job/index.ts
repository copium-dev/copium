export { default as Job } from './Job.svelte';

// If you have any types related to the Job component
export interface JobProps {
    id: string;
    company: string;
    role: string;
    appliedDate: Date;
    location: string;
    status?: string;
}