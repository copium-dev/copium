<script lang="ts">
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";

    import { Job } from "$lib/components/Job";
    import AddJob from "$lib/components/AddJob/AddJob.svelte";
    import FilterJobs from "$lib/components/FilterJobs/FilterJobs.svelte";
    import PaginateJobs from "$lib/components/PaginateJobs/PaginateJobs.svelte";

    import { filterStore } from "$lib/components/FilterJobs/filterStore";
    import { paginationStore } from "$lib/components/PaginateJobs/paginationStore";

    import type { PageData } from "./$types";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { buildParamsFromFilters } from "$lib/utils/filter";

    export let data: PageData;

    // subscribe to filter query store cause this is where
    // the raw texy query is edited
    $: query = $filterStore.query;

    // when loading a new page, ensure the total pages and count are updated
    onMount(() => {
        paginationStore.update((current) => ({
            ...current,
            count: 10 * data.totalPages,
        }));
    });

    // declare a variable to hold the FilterJobs component instance
    // we need this to extract the filter values from within the component
    let filterJobsComponent: FilterJobs;

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === "Enter") {
            e.preventDefault();

            const value = (e.currentTarget as HTMLInputElement).value;
            // update the store with the new query
            filterStore.update((current) => ({ ...current, query: value }));
            updateURL(filterJobsComponent); // updateURL will use query as well as the filter values
        }
    }

    function updateURL(filterJobsComponent: FilterJobs) {
        // read the filter values from the store
        const { query } = $filterStore;

        // Read the filter values from the FilterJobs component
        const { company, role, location, startDate, endDate, status } = filterJobsComponent;

        // Build the URL parameters
        const params = buildParamsFromFilters({ query, company, role, location, startDate, endDate, status });

        // Update the URL
        goto(`?${params.toString()}`);
    }
</script>

<div
    class="flex flex-col justify-start gap-4 items-stretch w-full sm:w-5/6 mx-auto h-full p-3 my-10"
>
    <div
        class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw]"
    >
        <div
            class="flex flex-row flex-grow gap-4 items-center w-full sm:w-auto"
        >
            <AddJob />
            <Separator orientation="vertical" class="h-6" />
            <Input
                type="text"
                placeholder="Search by company, role, or location"
                id="query"
                bind:value={query}
                on:keydown={handleKeyDown}
            />
            <Separator orientation="vertical" class="h-6" />
        </div>
        <!-- input must be in a different div to avoid problems so just use a store to apply the raw text query -->
        <FilterJobs bind:this={filterJobsComponent} />
    </div>

    <div class="mb-4">
        {#each data.applications as job (job.objectID)}
            <Job
                objectID={job.objectID}
                company={job.company}
                role={job.role}
                appliedDate={job.appliedDate}
                location={job.location}
                status={job.status}
                link={job.link}
            />
        {/each}
    </div>

    <PaginateJobs />
</div>