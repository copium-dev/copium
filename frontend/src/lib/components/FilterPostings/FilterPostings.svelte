<script lang="ts">
    import * as Popover from "$lib/components/ui/popover";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import { Label } from "$lib/components/ui/label";

    import { buildParamsFromFilters } from "$lib/utils/filter";

    import { postingsFilterStore } from "$lib/stores/postingsFilterStore";

    import { goto } from "$app/navigation"; 

    function updateFilterStore(e: Event) {
        const { id, value } = e.target as HTMLInputElement;
        postingsFilterStore.update(current => ({ ...current, [id]: value }));
    }

    function clearFilters() {
        postingsFilterStore.set({
            query: "",
            company: "",
            role: "",
            location: "",
            startDate: "",
            endDate: "",
        })
        // get hitsPerPage from URL before redirecting
        const url = new URL(window.location.href);
        const hitsPerPage = url.searchParams.get('hits');
        goto('/postings' + (hitsPerPage ? `?hits=${hitsPerPage}` : ''));
    }

    function sendFilters() {
        const { query, company, role, location, startDate, endDate } = $postingsFilterStore;
        const params = buildParamsFromFilters({ query, company, role, location, startDate, endDate });
        goto(`?${params.toString()}`); 
    }
</script>

<Popover.Root>
    <Popover.Trigger asChild let:builder>
        <Button builders={[builder]} variant="outline" id="filter"
            >Filter</Button
        >
    </Popover.Trigger>
    <Popover.Content class="w-56">
        <div class="grid gap-2">
            <div class="flex flex-col gap-2">
                <div>
                    <Label for="company" class="text-xs"
                        >Company</Label
                    >
                    <Input
                        type="text"
                        class="text-xs h-7"
                        id="company"
                        bind:value={$postingsFilterStore.company}
                        on:input={updateFilterStore}
                    />
                </div>
                <div>
                    <Label for="role" class="text-xs"
                        >Role</Label
                    >
                    <Input
                        type="text"
                        class="text-xs h-7"
                        id="role"
                        bind:value={$postingsFilterStore.role}
                        on:input={updateFilterStore}
                    />
                </div>
                <div>
                    <Label for="location" class="text-xs"
                        >Location</Label
                    >
                    <Input
                        type="text"
                        class="text-xs h-7"
                        id="location"
                        bind:value={$postingsFilterStore.location}
                        on:input={updateFilterStore}
                    />
                </div>
            </div>
            <div class="flex flex-col gap-2">
                <div>
                    <Label for="start-date" class="text-xs"
                        >Posted/Updated From</Label
                    >
                    <Input
                        type="date"
                        class="text-xs h-7"
                        id="start-date"
                        bind:value={$postingsFilterStore.startDate}
                        on:input={updateFilterStore}
                    />
                </div>
                <div>
                    <Label for="end-date" class="text-xs"
                        >Posted/Updated Until</Label
                    >
                    <Input
                        type="date"
                        class="text-xs h-7"
                        id="end-date"
                        bind:value={$postingsFilterStore.endDate}
                        on:input={updateFilterStore}
                    />
                </div>
            </div>
            <Separator orientation="horizontal" class="my-2" />
            <div class="flex justify-stretch gap-2">
                <Button
                    class="font-medium leading-none flex-grow"
                    on:click={sendFilters}
                >
                    Confirm
                </Button>
                <Button variant="outline"
                    on:click={clearFilters}
                    class="font-meidum leading-none flex-grow"
                    >Clear</Button
                >
            </div>
        </div>
    </Popover.Content>
</Popover.Root>