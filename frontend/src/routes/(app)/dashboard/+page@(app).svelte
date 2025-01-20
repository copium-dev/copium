<script lang="ts">
    export let data;

    // job component
    import Job from "./job.svelte";

    // mock data json file
    import mockData from "./mockdata.json";
    const jobs = mockData.jobs.map((job) => ({
        ...job,
        appliedDate: new Date(job.appliedDate),
    }));
</script>

<div class="flex flex-col justify-center gap-4 items-center h-full">
    <h1>Dashboard</h1>
    {#if data.email}
        <pre>{JSON.stringify(data.email, null, 2)}</pre>
        <div class="p-3 rounded-lg">
            {#each jobs as job (job.id)}
                <Job
                    id={job.id}
                    company={job.company}
                    role={job.role}
                    appliedDate={job.appliedDate}
                    location={job.location}
                    status={job.status}
                />
            {/each}
        </div>
    {:else}
        <p>Not logged in</p>
    {/if}
</div>
