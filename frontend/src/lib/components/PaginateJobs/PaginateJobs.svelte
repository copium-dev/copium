<script>
    import * as Pagination from "$lib/components/ui/pagination";

    import { paginationStore } from "$lib/components/PaginateJobs/paginationStore";

    import { changePage } from "$lib/utils/filter";

    import { goto } from "$app/navigation";

    $: count = $paginationStore.count;

    const perPage = 10;
    const siblingCount = 1;

    function nextPage() {
        const params = changePage("next");
        goto(`?${params.toString()}`);
    }

    function prevPage() {
        const params = changePage("prev");
        goto(`?${params.toString()}`);
    }
</script>

<Pagination.Root {count} {perPage} {siblingCount} let:pages let:currentPage>
    <Pagination.Content>
        <Pagination.Item>
            <Pagination.PrevButton on:click={prevPage}>
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
                        isActive={currentPage === page.value}
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
            </Pagination.NextButton>
        </Pagination.Item>
    </Pagination.Content>
</Pagination.Root>
