// to be used for job status reverts
import { writable } from "svelte/store";

export const statusUpdateStore = writable({
    ok: false,
    jobID: '',
    role: '',
    company: '',
    status: '',
    prevStatus: '',
})