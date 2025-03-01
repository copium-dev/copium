<script>
    import * as Pagination from "$lib/components/ui/pagination";
    import ChevronLeft from "svelte-radix/ChevronLeft.svelte";
    import ChevronRight from "svelte-radix/ChevronRight.svelte";

    import { postingsPaginationStore } from "$lib/stores/postingsPaginationStore";

    import { changePage } from "$lib/utils/filter";

    import { goto } from "$app/navigation";

    $: count = $postingsPaginationStore.count;

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
                <ChevronRight class="h-4 w-4"/>
            </Pagination.NextButton>
        </Pagination.Item>
    </Pagination.Content>
</Pagination.Root>
