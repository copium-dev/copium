<script lang="ts">
    export let data;
    import { Button } from "$lib/components/ui/button";
    import Fa from "svelte-fa";

    // job component
    import { Job } from "$lib/components/Job";

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
        <div class="w-fit p-3">
            <div class="flex flex-row justify-end w-full">
                <Button variant="outline" class="w-16">
                    <!-- <Fa icon={} />  -->
                    Add
                </Button>
            </div>
            <div class="rounded-lg">
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
        </div>
    {:else}
        <p>Not logged in</p>
    {/if}
</div>
