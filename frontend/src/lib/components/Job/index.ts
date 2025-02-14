export { default as Job } from './Job.svelte';

export interface JobProps {
    objectID: string;
    company: string;
    role: string;
    appliedDate: Date;
    location: string;
    status: string;
    link: string | undefined | null;
}