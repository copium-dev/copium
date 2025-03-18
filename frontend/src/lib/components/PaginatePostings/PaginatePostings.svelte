<script lang="ts">
    import * as Pagination from "$lib/components/ui/pagination";
    import { Separator } from "$lib/components/ui/separator";
    import ChevronLeft from "svelte-radix/ChevronLeft.svelte";
    import ChevronRight from "svelte-radix/ChevronRight.svelte";

    import { postingsPaginationStore } from "$lib/stores/postingsPaginationStore";

    import { changePage } from "$lib/utils/filter";

    import { page as pageStore } from "$app/state";
    import { afterNavigate } from "$app/navigation";
    import { goto } from "$app/navigation";
    
    import Input from "$lib/components/ui/input/input.svelte";

    $: count = $postingsPaginationStore.count;
    $: perPage = $postingsPaginationStore.perPage;

    const siblingCount = 1;

    // get current page from URL because it needs to follow the URL for correct pagination
    let currentPageFromURL = parseInt(
        pageStore.url.searchParams.get("page") || "1"
    );

    $: postingsPaginationStore.update((state) => ({
        ...state,
        currentPage: currentPageFromURL ? currentPageFromURL : 1,
    }));

    afterNavigate(() => {
        currentPageFromURL = parseInt(
            pageStore.url.searchParams.get("page") || "1"
        );
    });

    function nextPage() {
        const params = changePage("next");
        goto(`?${params.toString()}`);
    }

    function prevPage() {
        const params = changePage("prev");
        goto(`?${params.toString()}`);
    }

    let pageValue = "";

    function goToPage(event: Event) {
        event.preventDefault();

        let pageNum = parseInt(pageValue) || 1;
        const params = new URLSearchParams(pageStore.url.search);

        if (pageNum < 1) {
            pageNum = 1;
        }

        params.set("page", pageNum.toString());
        goto(`?${params.toString()}`);

        pageValue = "";
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            goToPage(event);
        }
    }
</script>

<Pagination.Root
    {count}
    {perPage}
    {siblingCount}
    page={currentPageFromURL}
    let:pages
    class="flex flex-row w-fit mx-0"
>
    <div class="flex flex-row items-center gap-2">
        <div class="text-sm">Go to page:</div>
            <form on:submit={goToPage} class="flex items-center">
                <Input
                    type="text"
                    bind:value={pageValue}
                    on:keydown={handleKeydown}
                    placeholder={(currentPageFromURL || 1).toString()}
                    class="focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 w-[50px]"
                />
            </form>
            <Separator orientation="vertical" class="mx-3 h-6" />
    </div>
    <Pagination.Content>
        <Pagination.Item>
            <Pagination.PrevButton on:click={prevPage}>
                <ChevronLeft class="h-4 w-4" />
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
                <ChevronRight class="h-4 w-4" />
            </Pagination.NextButton>
        </Pagination.Item>
    </Pagination.Content>
</Pagination.Root>
