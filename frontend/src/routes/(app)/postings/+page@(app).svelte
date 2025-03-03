<script lang="ts">
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import * as Table from "$lib/components/ui/table/index.js";
    import { buttonVariants } from "$lib/components/ui/button";

    import { Map, Calendar, Building2, Link, BriefcaseBusiness }  from "lucide-svelte";

    import placeholder from "$lib/images/placeholder.png";

    import { buildParamsFromFilters } from "$lib/utils/filter";
    import FilterPostings from "$lib/components/FilterPostings/FilterPostings.svelte";
    import { postingsFilterStore } from "$lib/stores/postingsFilterStore";

    import PaginatePostings from "$lib/components/PaginatePostings/PaginatePostings.svelte";
    import { postingsPaginationStore } from "$lib/stores/postingsPaginationStore";

    import type { PageData } from "./$types";
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";

    export let data: PageData;

    // reactive block to update pagination count
    //  - onMount does not work here since page data is updated but
    //    component is not rerendered when goto() is called
    $: postingsPaginationStore.update((current) => ({
        ...current,
        count: 10 * data.totalPages,
    }));

    function updateInput(e: Event) {
        const value = (e.currentTarget as HTMLInputElement).value;
        postingsFilterStore.update((current) => ({ ...current, query: value }));
    }

    // only if input is focused, handle Enter
    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === "Enter") {
            const queryElement = document.getElementById(
                "query",
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
        const { query, company, role, location, startDate, endDate } =
            $postingsFilterStore;
        const params = buildParamsFromFilters({
            query,
            company,
            role,
            location,
            startDate,
            endDate,
        });
        goto(`?${params.toString()}`);
    }

    function formatDate(timestamp: number): string {
        if (!timestamp) return "";

        const date = new Date(timestamp);
        if (isNaN(date.getTime())) return "";

        // Adjust for timezone
        const adjustedDate = new Date(
            date.getTime() + date.getTimezoneOffset() * 60 * 1000
        );

        const month = String(adjustedDate.getMonth() + 1).padStart(2, "0");
        const day = String(adjustedDate.getDate()).padStart(2, "0");
        const year = adjustedDate.getFullYear();

        return `${month}-${day}-${year}`;
    }

</script>

<div
    class="flex flex-col justify-start gap-4 items-stretch w-full my-12"
>
    <!-- sticky controls wrapper -->
    <div class="sticky bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 z-50">
        <div
            class="bg-background flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw] py-2"
        >
            <div class="flex flex-col w-full">
                <div
                    class="flex flex-row gap-4 items-center w-full sm:w-auto mb-4"
                >
                    <!-- NOTE: MUST PUT bind:value={$postingsFilterStore.query} -->
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
               <PaginatePostings />
            </div>
        </div>
    </div>

    <Table.Root class="mb-4">
        <Table.Header>
            <Table.Row>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Company
                        <Building2 class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Role
                        <BriefcaseBusiness class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Locations
                        <Map class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Posted
                        <Calendar class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Updated
                        <Calendar class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
                <Table.Head>
                    <span class="inline-flex items-center gap-2">
                        Link
                        <Link class="w-[15px] h-[15px] stroke-[1.5]" />
                    </span>
                </Table.Head>
            </Table.Row>
        </Table.Header>
        <Table.Body>
            {#each data.postings as posting, i (i)}
                <Table.Row>
                    <Table.Cell class="inline-flex items-center gap-2 h-12">
                        <img 
                            src={data.companyLogos[posting.company_name] || placeholder}
                            alt={posting.company_name}
                            class="w-6 h-6 rounded-lg object-cover"
                        />
                        {posting.company_name}
                    </Table.Cell>
                    <Table.Cell>{posting.title}</Table.Cell>
                    <Table.Cell>{posting.locations?.join(' | ') || ''}</Table.Cell>
                    <Table.Cell>{formatDate(posting.date_posted)}</Table.Cell>
                    <Table.Cell>{formatDate(posting.date_updated)}</Table.Cell>
                    <Table.Cell>
                        <a
                            href={posting.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            class={buttonVariants({size: "sm"})}
                        >
                            Apply
                        </a>
                    </Table.Cell>
                </Table.Row>
            {/each}
        </Table.Body>
    </Table.Root>
</div>