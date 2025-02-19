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
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { buildParamsFromFilters } from "$lib/utils/filter";

    export let data: PageData;

    // reactive block to update pagination count
    //  - onMount does not work here since page data is updated but
    //    component is not rerendered when goto() is called
    $: paginationStore.update((current) => ({
        ...current,
        count: 10 * data.totalPages,
    }));

    function updateInput(e: Event) {
        const value = (e.currentTarget as HTMLInputElement).value;
        filterStore.update((current) => ({ ...current, query: value }));
    }

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === "Enter") {
            e.preventDefault();

            // update the store with the new query
            updateInput(e);
            updateURL(); // updateURL will use raw text query AND filter values from FilterJobs component
        }
    }

    // shortcuts for search input
    function handleGlobalKeydown(e: KeyboardEvent) {
        if (e.key === "f") {
            e.preventDefault();
            const queryElement = document.getElementById(
                "query",
            ) as HTMLInputElement | null;
            if (queryElement) queryElement.focus();
        }
        if (e.key == "Escape") {
            const queryElement = document.getElementById(
                "query",
            ) as HTMLInputElement | null;
            if (queryElement) queryElement.blur();
        }
    }

    onMount(() => {
        if (typeof window !== "undefined") {
            window.addEventListener("keydown", handleGlobalKeydown);
        }
    });

    onDestroy(() => {
        if (typeof window !== "undefined") {
            window.removeEventListener("keydown", handleGlobalKeydown);
        }
    });

    function updateURL() {
        const { query, company, role, location, startDate, endDate, status } =
            $filterStore;
        const params = buildParamsFromFilters({
            query,
            company,
            role,
            location,
            startDate,
            endDate,
            status,
        });
        goto(`?${params.toString()}`);
    }
</script>

<div
    class="flex flex-col justify-start gap-4 items-stretch w-full sm:w-5/6 mx-auto min-h-full p-3 my-5"
>
    <!-- sticky controls wrapper -->
    <div class="sticky top-0 bg-background z-10">
        <div
            class="bg-background flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw] py-2"
        >
            <div class="flex flex-col w-full">
                <div
                    class="flex flex-row gap-4 items-center w-full sm:w-auto mb-4"
                >
                    <AddJob />
                    <Separator orientation="vertical" class="h-6" />
                    <Input
                        type="text"
                        placeholder="Press ENTER to search by company, role, or location."
                        id="query"
                        bind:value={$filterStore.query}
                        on:input={updateInput}
                        on:keydown={handleKeyDown}
                    />
                    <Separator orientation="vertical" class="h-6" />
                    <FilterJobs />
                </div>
                <PaginateJobs />
            </div>
        </div>
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
</div>
