<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import { Job } from "$lib/components/Job";

    import { enhance } from "$app/forms";
    import type { PageData } from './$types';
    
    let isAddOpen = false;

    function toggleAdd() {
        isAddOpen = !isAddOpen;
    }

    export let data: PageData;

</script>

<div class="flex flex-col justify-start gap-4 items-center h-full">
    <div class="w-full sm:w-5/6 p-3 my-10 ">
        <div class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw]">
            <div class="flex flex-row flex-grow gap-4 items-center w-full sm:w-auto">
                <Button variant="outline" class="w-16" on:click={toggleAdd}>
                    Add
                </Button>
                {#if isAddOpen}
                    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
                        <div class="relative flex flex-col gap-2 my-2 inline-block w-full overflow-visible rounded-2xl bg-white dark:bg-zinc-900 px-4 py-5 sm:p-6 text-left align-bottom shadow-xl transition-all sm:align-middle sm:max-w-3xl">
                            <h1 class="text-xl">Add Application</h1>
                            <form
                                action="?/add"
                                method="POST"
                                class="flex flex-col gap-2 w-full"
                                use:enhance={() => {
                                    toggleAdd();
                                    return ({ update }) => {
                                        update();
                                    };
                                }}
                            >
                                <Input type="text" name="role" placeholder="Role" required />
                                <Input type="text" name="company" placeholder="Company" required />
                                <Input type="text" name="location" placeholder="Location" required />
                                <Input type="text" name="link" placeholder="Link (Optional)" />
                                <Input type="date" name="appliedDate" placeholder="Applied Date" required />
                                <div class="flex flex-row gap-2 items-stretch justify-between w-full">
                                    <Button type="button" variant="outline" class="flex-grow text-red-500 hover:text-red-500" on:click={toggleAdd}>
                                        Cancel
                                    </Button>
                                    <Button type="submit" variant="outline" class="flex-grow">
                                        Add
                                    </Button>
                                </div>
                            </form>
                        </div>
                    </div>
                {/if}
                <Separator orientation="vertical" class="h-6" />
                <Input type="text" placeholder="Search for roles or companies" />
            </div>
            <div class="flex flex-row gap-2 sm:gap-4 items-center w-full sm:w-auto">
                <div class="flex items-center gap-0 sm:gap-2 -ml-3 sm:ml-0">
                    <Button variant="ghost" type="button" class="text-xs sm:text-sm">Applied from</Button>
                    <Separator orientation="horizontal" class="w-2 sm:w-3" />
                    <Button variant="ghost" type="button" class="text-xs sm:text-sm">Applied until</Button>
                </div>
                <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" type="button" class="text-xs sm:text-sm">Job Type</Button>
                <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" type="button" class="text-xs sm:text-sm">Status</Button>
            </div>
        </div>

        <div>
            {#each data.applications as job (job.id)}
                <Job
                    id={job.id}
                    company={job.company}
                    role={job.role}
                    appliedDate={job.appliedDate}
                    location={job.location}
                    status={job.status}
                    link={job.link}
                />
            {/each}
        </div>
    </div>
</div>