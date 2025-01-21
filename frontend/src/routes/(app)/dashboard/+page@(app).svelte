<script lang="ts">
    export let data;
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";

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
    <!-- <h1>Dashboard</h1> -->
    <pre>{JSON.stringify(data.email, null, 2)}</pre>

    <div class="w-fit p-3 my-10">
        <div
            class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full"
        >
            <div class="flex flex-row flex-grow gap-4 items-center w-full sm:w-auto">
                <Button variant="outline" class="w-16">Add</Button>
                <Separator orientation="vertical" class="h-6" />
                <Input type="email" placeholder="Search for roles or companies" />
            </div>
            <div class="flex flex-row gap-2 sm:gap-4 items-center w-full sm:w-auto">
                <div class="flex items-center gap-0 sm:gap-2">
                    <Button variant="ghost" type="submit" class="text-xs sm:text-sm"
                        >Applied from</Button
                    >
                    <Separator orientation="horizontal" class="w-2 sm:w-3" />
                    <Button variant="ghost" type="submit" class="text-xs sm:text-sm"
                        >Applied until</Button
                    >
                </div>
                <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" type="submit" class="text-xs sm:text-sm"
                    >Job Type</Button
                >
                <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" type="submit" class="text-xs sm:text-sm"
                    >Status</Button
                >
            </div>
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
</div>
