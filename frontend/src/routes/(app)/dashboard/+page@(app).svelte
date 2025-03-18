<script lang="ts">
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import Switch from "$lib/components/ui/switch/switch.svelte";

    import { Job } from "$lib/components/Job";
    import GridJob from "$lib/components/GridJob/GridJob.svelte";
    import AddJob from "$lib/components/AddJob/AddJob.svelte";
    import FilterJobs from "$lib/components/FilterJobs/FilterJobs.svelte";
    import PaginateJobs from "$lib/components/PaginateJobs/PaginateJobs.svelte";

    import { jobsFilterStore } from "$lib/stores/jobsFilterStore";
    import { dashboardPaginationStore } from "$lib/stores/dashboardPaginationStore";

    import { List, LayoutGrid } from "lucide-svelte";

    import type { PageData } from "./$types";
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { buildParamsFromFilters } from "$lib/utils/filter";
    import { browser } from "$app/environment";

    export let data: PageData;
    export let form; // for eager loading

    $: if (form?.type === "success" && form?.data) {
        // in svelte, reactive updates triggered by assignments not mutations
        // so we gotta do all this below instead of an unshift
        data = {
            ...data,
            applications: [
                {
                    objectID: form.data.objectID,
                    company: form.data.company,
                    role: form.data.role,
                    appliedDate: form.data.appliedDate,
                    location: form.data.location,
                    status: form.data.status,
                    link: form.data.link,
                },
                ...data.applications,
            ],
        };
    }

    // reactive block to update pagination count
    //  - onMount does not work here since page data is updated but
    //    component is not rerendered when goto() is called
    $: dashboardPaginationStore.update((current) => ({
        ...current,
        count: 10 * data.totalPages,
    }));

    function updateInput(e: Event) {
        const value = (e.currentTarget as HTMLInputElement).value;
        jobsFilterStore.update((current) => ({ ...current, query: value }));
    }

    // only if input is focused, handle Enter
    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === "Enter") {
            const queryElement = document.getElementById(
                "query"
            ) as HTMLInputElement | null;
            if (document.activeElement !== queryElement) return;
            e.preventDefault();

            // update the store with the new query
            updateInput(e);
            updateURL(); // updateURL will use raw text query AND filter values from FilterJobs component
        }
    }

    // shortcuts for search input; only works when body is focused
    function handleGlobalKeydown(e: KeyboardEvent) {
        if (e.key == "Escape") {
            const queryElement = document.getElementById(
                "query"
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

    // list - grid style toggle
    let isViewPreferenceLoaded = false;
    let isGridView: boolean | undefined = undefined;

    function updateURL() {
        const { query, company, role, location, startDate, endDate, status } =
            $jobsFilterStore;
        const params = buildParamsFromFilters({
            query,
            company,
            role,
            location,
            startDate,
            endDate,
            status,
            // we can load more in grid view
            hitsPerPage: isGridView ? 20 : 10,
        });
        goto(`?${params.toString()}`);
    }

    // Use a reactive statement that runs as soon as possible client-side
    $: if (browser && isGridView === undefined) {
        const savedView = localStorage.getItem("view_preference_dashboard");
        isGridView = savedView === "true";
        isViewPreferenceLoaded = true;
    }

    onMount(() => {
        window.addEventListener("keydown", handleGlobalKeydown);

        return () => {
            window.removeEventListener("keydown", handleGlobalKeydown);
        };
    });

    // save view preference
    $: if (browser && isViewPreferenceLoaded && isGridView !== undefined) {
        localStorage.setItem("view_preference_dashboard", isGridView.toString())
        updateURL();    // must reload with saved view preference because of hitsPerPage
    }
</script>

<div class="flex flex-col justify-start gap-4 items-stretch w-full my-12">
    <!-- sticky controls wrapper -->
    <div
        class="px-8 sticky bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 z-50"
    >
        <div
            class="bg-background flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full py-2"
        >
            <div class="flex flex-col w-full">
                <div
                    class="flex flex-row gap-4 items-center w-full sm:w-auto mb-4"
                >
                    <AddJob />
                    <Separator orientation="vertical" class="h-6" />
                    <Input
                        type="text"
                        placeholder="Search by company, role, or location."
                        id="query"
                        bind:value={$jobsFilterStore.query}
                        on:input={updateInput}
                        on:keydown={handleKeyDown}
                    />
                    <Separator orientation="vertical" class="h-6" />
                    <FilterJobs />
                </div>
                {#if isViewPreferenceLoaded}
                    <div
                        class="flex flex-row gap-4 justify-between items-center w-full sm:w-auto"
                    >
                        <div class="flex gap-2 items-center justify-center">
                            <div
                                class={!isGridView
                                    ? "flex items-center gap-1 text-sm font-medium"
                                    : "flex items-center gap-1 text-muted-foreground text-sm font-medium"}
                            >
                                <List class="w-[15px] h-[17px] stroke-[1.5]" />
                                List
                            </div>
                            <Switch bind:checked={isGridView} />
                            <div
                                class={isGridView
                                    ? "flex items-center gap-1 text-sm font-medium"
                                    : "flex items-center gap-1 text-muted-foreground text-sm font-medium"}
                            >
                                <LayoutGrid
                                    class="w-[15px] h-[17px] stroke-[1.5]"
                                />
                                Grid
                            </div>
                        </div>
                        <PaginateJobs />
                    </div>
                {/if}
            </div>
        </div>
    </div>

    {#if isViewPreferenceLoaded}
        {#if !isGridView}
            <div class="mb-4">
                <!-- by default, visible is true. but for eager loading, if delete application called within Job
                visible is set to false and there is a if block to only render if the job is visible -->
                {#each data.applications as job (job.objectID)}
                    <Job
                        objectID={job.objectID}
                        company={job.company}
                        role={job.role}
                        appliedDate={job.appliedDate}
                        location={job.location}
                        status={job.status}
                        link={job.link}
                        visible={true}
                    />
                {/each}
            </div>
        {:else}
            <div class="flex flex-wrap items-center justify-start grid-rows-auto px-6 mb-4">
                {#each data.applications as job (job.objectID)}
                    <GridJob
                        objectID={job.objectID}
                        company={job.company}
                        role={job.role}
                        appliedDate={job.appliedDate}
                        location={job.location}
                        status={job.status}
                        link={job.link}
                        visible={true}
                    />
                {/each}
            </div>
        {/if}
    {/if}

    
</div>
