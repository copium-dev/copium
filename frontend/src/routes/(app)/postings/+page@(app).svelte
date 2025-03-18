<script lang="ts">
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import * as Table from "$lib/components/ui/table/index.js";
    import * as HoverCard from "$lib/components/ui/hover-card/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import { Button } from "$lib/components/ui/button";
    import Switch from "$lib/components/ui/switch/switch.svelte";

    import {
        Map,
        Calendar,
        Building2,
        Link,
        BriefcaseBusiness,
        List,
        LayoutGrid,
    } from "lucide-svelte";

    import placeholder from "$lib/images/placeholder.png";

    import { formatDateForDisplay, formatDateForInput } from "$lib/utils/date";

    import { buildParamsFromFilters } from "$lib/utils/filter";
    import FilterPostings from "$lib/components/FilterPostings/FilterPostings.svelte";
    import { postingsFilterStore } from "$lib/stores/postingsFilterStore";

    import PaginatePostings from "$lib/components/PaginatePostings/PaginatePostings.svelte";
    import { postingsPaginationStore } from "$lib/stores/postingsPaginationStore";

    import type { PageData } from "./$types";
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { browser } from "$app/environment";

    export let data: PageData;

    // reactive block to update pagination count
    //  - onMount does not work here since page data is updated but
    //    component is not rerendered when goto() is called
    $: postingsPaginationStore.update((current) => ({
        ...current,
        count: isGridView ? 20 * data.totalPages : 10 * data.totalPages,
        perPage: isGridView ? 20 : 10,
    }));

    function updateInput(e: Event) {
        const value = (e.currentTarget as HTMLInputElement).value;
        postingsFilterStore.update((current) => ({ ...current, query: value }));
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
        const { query, company, role, location, startDate, endDate } =
            $postingsFilterStore;
        const params = buildParamsFromFilters({
            query,
            company,
            role,
            location,
            startDate,
            endDate,
            // we can load more in grid view
            hitsPerPage: isGridView ? 20 : 10,
        });

        // this has to be updated in updateURL as well because user may have changed view preference
        postingsPaginationStore.update((current) => ({
            ...current,
            count: isGridView ? 20 * data.totalPages : 10 * data.totalPages,
            perPage: isGridView ? 20 : 10,
        }));

        goto(`?${params.toString()}`);
    }

    // Use a reactive statement that runs as soon as possible client-side
    $: if (browser && isGridView === undefined) {
        const savedView = localStorage.getItem("view_preference_postings");
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
        localStorage.setItem("view_preference_postings", isGridView.toString());
        updateURL(); // must reload with saved view preference because of hitsPerPage
    }

    function addApplication(posting: any) {
        const company = posting.company_name;
        const role = posting.title;
        const location = posting.locations[0];
        const link = posting.url;
        const appliedDate = formatDateForInput(Math.floor(Date.now() / 1000));

        const form = new FormData();
        form.append("company", company);
        form.append("role", role);
        form.append("location", location);
        form.append("link", link);
        form.append("appliedDate", appliedDate.toString());

        // dont need to wait for response just fire and forget
        fetch("dashboard?/add", {
            method: "POST",
            body: form,
        });
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
                    <Input
                        type="text"
                        placeholder="Search by company, role, or location."
                        id="query"
                        on:input={updateInput}
                        on:keydown={handleKeyDown}
                    />
                    <Separator orientation="vertical" class="h-6" />
                    <FilterPostings />
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
                        <PaginatePostings />
                    </div>
                {/if}
            </div>
        </div>
    </div>

    {#if isViewPreferenceLoaded}
        {#if !isGridView}
            <Table.Root class="overflow-hidden table-fixed">
                <Table.Header>
                    <Table.Row class="border-b border-dashed">
                        <Table.Head class="border-r border-dashed pl-8 w-3/12">
                            <span class="inline-flex items-center gap-2">
                                <Building2
                                    class="w-[15px] h-[17px] stroke-[1.5]"
                                />
                                Company
                            </span>
                        </Table.Head>
                        <Table.Head class="border-r border-dashed w-5/12">
                            <span class="inline-flex items-center gap-2">
                                <BriefcaseBusiness
                                    class="w-[15px] h-[17px] stroke-[1.5]"
                                />
                                Role
                            </span>
                        </Table.Head>
                        <Table.Head class="border-r border-dashed w-2/12">
                            <span class="inline-flex items-center gap-2">
                                <Map class="w-[15px] h-[17px] stroke-[1.5]" />
                                Locations
                            </span>
                        </Table.Head>
                        <Table.Head class="border-r border-dashed w-2/12">
                            <span class="inline-flex items-center gap-2">
                                <Calendar
                                    class="w-[15px] h-[17px] stroke-[1.5]"
                                />
                                Posted
                            </span>
                        </Table.Head>
                        <Table.Head class="border-r border-dashed w-2/12">
                            <span class="inline-flex items-center gap-2">
                                <Calendar
                                    class="w-[15px] h-[17px] stroke-[1.5]"
                                />
                                Updated
                            </span>
                        </Table.Head>
                        <Table.Head class="pr-8 w-1/12">
                            <span class="inline-flex items-center gap-2">
                                <Link class="w-[15px] h-[17px] stroke-[1.5]" />
                                Link
                            </span>
                        </Table.Head>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {#each data.postings as posting, i (i)}
                        <Table.Row
                            class="border-b border-dashed dark:brightness-[0.9]"
                        >
                            <Table.Cell
                                class="border-r border-dashed w-full inline-flex items-center gap-2 h-12 pl-8"
                            >
                                <img
                                    src={data.companyLogos[
                                        posting.company_name
                                    ] || placeholder}
                                    alt={posting.company_name}
                                    class="w-6 h-6 rounded-lg object-cover"
                                />
                                <p class="truncate">
                                    {posting.company_name}
                                </p>
                            </Table.Cell>
                            <Table.Cell class="border-r border-dashed">
                                <p class="truncate">
                                    {posting.title}
                                </p>
                            </Table.Cell>
                            <Table.Cell class="border-r border-dashed">
                                <HoverCard.Root>
                                    <HoverCard.Trigger
                                        class="rounded-sm underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-8 focus-visible:outline-black"
                                    >
                                        <p class="truncate">
                                            {#if posting.locations.length > 1}
                                                {posting.locations[0]}
                                                <span class="font-semibold"
                                                    >+{posting.locations
                                                        .length - 1}</span
                                                >
                                            {:else}
                                                {posting.locations[0]}
                                            {/if}
                                        </p>
                                    </HoverCard.Trigger>
                                    <HoverCard.Content class="w-fit">
                                        <div
                                            class="flex justify-between space-x-4"
                                        >
                                            <div class="space-y-1">
                                                {#each posting.locations as location}
                                                    <p class="text-sm">
                                                        {location}
                                                    </p>
                                                {/each}
                                            </div>
                                        </div>
                                    </HoverCard.Content>
                                </HoverCard.Root>
                            </Table.Cell>
                            <Table.Cell class="border-r border-dashed"
                                ><p class="truncate">
                                    {formatDateForDisplay(posting.date_posted)}
                                </p></Table.Cell
                            >
                            <Table.Cell class="border-r border-dashed"
                                ><p class="truncate">
                                    {formatDateForDisplay(posting.date_updated)}
                                </p></Table.Cell
                            >
                            <Table.Cell
                                class="flex items-center justify-center pr-8"
                            >
                                <AlertDialog.Root>
                                    <AlertDialog.Trigger asChild let:builder>
                                        <Button
                                            builders={[builder]}
                                            href={posting.url}
                                            target="_blank"
                                            size="sm"
                                        >
                                            Apply
                                        </Button>
                                    </AlertDialog.Trigger>
                                    <AlertDialog.Content>
                                        <AlertDialog.Header>
                                            <AlertDialog.Title>
                                                Did you apply for this job?
                                            </AlertDialog.Title>
                                            <AlertDialog.Description>
                                                Click "Yes" below to
                                                automatically add this to your
                                                dashboard.
                                            </AlertDialog.Description>
                                        </AlertDialog.Header>
                                        <AlertDialog.Footer>
                                            <AlertDialog.Cancel
                                                >No</AlertDialog.Cancel
                                            >
                                            <AlertDialog.Action
                                                on:click={() =>
                                                    addApplication(posting)}
                                            >
                                                Yes
                                            </AlertDialog.Action>
                                        </AlertDialog.Footer>
                                    </AlertDialog.Content>
                                </AlertDialog.Root>
                            </Table.Cell>
                        </Table.Row>
                    {/each}
                </Table.Body>
            </Table.Root>
        {:else}
            <!-- Grid View -->
            <div
                class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4 px-8"
            >
                {#each data.postings as posting, i (i)}
                    <div
                        class="border rounded-lg p-4 flex flex-col gap-1 h-full dark:brightness-[0.9]"
                    >
                        <div class="flex items-center gap-2 mb-2">
                            <img
                                src={data.companyLogos[posting.company_name] ||
                                    placeholder}
                                alt={posting.company_name}
                                class="w-8 h-8 rounded-lg object-cover"
                            />
                            <h3 class="font-medium truncate flex-1">
                                {posting.company_name}
                            </h3>

                            <div class="mt-auto">
                                <Button
                                    href={posting.url}
                                    target="_blank"
                                    size="sm"
                                    class="w-full px-2">Apply</Button
                                >
                            </div>
                        </div>

                        <h4 class="text-sm font-medium truncate">
                            {posting.title}
                        </h4>

                        <div
                            class="text-xs text-muted-foreground flex flex-col items-start"
                        >
                            
                            <HoverCard.Root>
                                <HoverCard.Trigger
                                    class="flex gap-1 items-center rounded-sm underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-8 focus-visible:outline-black"
                                >
                                <Map class="w-3 h-3 inline" />
                                    <p class="truncate">
                                        {#if posting.locations.length > 1}
                                            {posting.locations[0]}
                                            <span class="font-semibold"
                                                >+{posting.locations.length -
                                                    1}</span
                                            >
                                        {:else}
                                            {posting.locations[0]}
                                        {/if}
                                    </p>
                                </HoverCard.Trigger>
                                <HoverCard.Content class="w-fit">
                                    <div class="flex justify-between space-x-4">
                                        <div class="space-y-1">
                                            {#each posting.locations as location}
                                                <p class="text-sm">
                                                    {location}
                                                </p>
                                            {/each}
                                        </div>
                                    </div>
                                </HoverCard.Content>
                            </HoverCard.Root>
                            <div
                                class="text-xs text-muted-foreground flex items-center gap-1"
                            >
                                <Calendar class="w-3 h-3 inline" />
                                <span
                                    >Posted: {formatDateForDisplay(
                                        posting.date_posted
                                    )}</span
                                >
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    {/if}

    <div class="container flex justify-end gap-4">
        <p class="text-muted-foreground text-xs">*Only shows active postings</p>
        <!-- <p class="text-muted-foreground text-xs">*Updated every 3 hours</p> -->
        <div class="flex space-x-1">
            <p class="text-muted-foreground text-xs">
                *Internship postings from
            </p>
            <a
                href="https://github.com/cvrve/Summer2025-Internships"
                class="text-muted-foreground text-xs">[cvrve]</a
            >
        </div>
    </div>
</div>
