<script>
    import * as Pagination from "$lib/components/ui/pagination";
    import ChevronLeft from "svelte-radix/ChevronLeft.svelte";
    import ChevronRight from "svelte-radix/ChevronRight.svelte";

    import { postingsPaginationStore } from "$lib/stores/postingsPaginationStore";

    import { changePage } from "$lib/utils/filter";
    import { page } from '$app/stores';
    import { goto } from "$app/navigation";

    $: count = $postingsPaginationStore.count;

    const perPage = 10;
    const siblingCount = 1;

    // get current page from URL because it needs to follow the URL for correct pagination
    $: currentPageFromURL = parseInt($page.url.searchParams.get('page') || '1');

    // update the store with the current page from URL
    $: {
        const pageFromURL = parseInt($page.url.searchParams.get('page') || '1');
        postingsPaginationStore.update(state => ({
            ...state,
            currentPage: pageFromURL
        }));
    }

    function nextPage() {
        const params = changePage("next");
        goto(`?${params.toString()}`);
    }

    function prevPage() {
        const params = changePage("prev");
        goto(`?${params.toString()}`);
    }
</script>

<Pagination.Root {count} {perPage} {siblingCount} page={currentPageFromURL} let:pages>
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
