<script>
    import * as Pagination from "$lib/components/ui/pagination";
    import ChevronLeft from "svelte-radix/ChevronLeft.svelte";
    import ChevronRight from "svelte-radix/ChevronRight.svelte";

    import { dashboardPaginationStore } from "$lib/stores/dashboardPaginationStore";

    import { changePage } from "$lib/utils/filter";

    import { goto } from "$app/navigation";
    import { afterNavigate } from "$app/navigation";
    import { page as pageStore } from "$app/state";

    $: count = $dashboardPaginationStore.count;

    const perPage = 10;
    const siblingCount = 1;

    let currentPageFromURL = parseInt(pageStore.url.searchParams.get('page') || '1');

    $: dashboardPaginationStore.update(state => ({
        ...state,
        currentPage: currentPageFromURL
    }));

    afterNavigate(() => {
        currentPageFromURL = parseInt(pageStore.url.searchParams.get('page') || '1');
    });

    function nextPage() {
        const params = changePage("next");
        goto(`?${params.toString()}`);
    }

    function prevPage() {
        const params = changePage("prev");
        goto(`?${params.toString()}`);
    }
</script>

<Pagination.Root {count} {perPage} {siblingCount} let:pages page={currentPageFromURL}>
    <Pagination.Content>
        <Pagination.Item>
            <Pagination.PrevButton on:click={prevPage}>
                <ChevronLeft class="h-4 w-4"/>
                <span class="hidden sm:block">Prev</span>
            </Pagination.PrevButton>
        </Pagination.Item>
        {#each pages as page (page.key)}
            {#if page.type === "ellipsis"}
                <Pagination.Item>
                    <Pagination.Ellipsis />
                </Pagination.Item>
            {:else}
                <Pagination.Item>
                    <Pagination.Link
                        {page}
                        isActive={currentPageFromURL === page.value}
                        on:click={() => {
                            const params = changePage(page.value);
                            goto(`?${params.toString()}`);
                        }}
                    >
                        {page.value}
                    </Pagination.Link>
                </Pagination.Item>
            {/if}
        {/each}
        <Pagination.Item>
            <Pagination.NextButton on:click={nextPage}>
                <span class="hidden sm:block">Next</span>
                <ChevronRight class="h-4 w-4"/>
            </Pagination.NextButton>
        </Pagination.Item>
    </Pagination.Content>
</Pagination.Root>
