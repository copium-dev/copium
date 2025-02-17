<script lang="ts">
    import * as Popover from "$lib/components/ui/popover";
    import * as Select from "$lib/components/ui/select";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import { Label } from "$lib/components/ui/label";

    import { buildParamsFromFilters } from "$lib/utils/filter";

    import { filterStore } from "$lib/components/FilterJobs/filterStore";

    import { goto } from "$app/navigation";

    // filter state
    export let company = ""; // Export the filter values
    export let role = "";
    export let location = "";
    export let startDate = "";
    export let endDate = "";
    export let status = "Status";

    $: query = $filterStore.query;

    function changeStatus(newStatus: string) {
        status = newStatus;
    }
    
    function clearFilters() {
        filterStore.update(current => ({ ...current, query: "" }));
        company = "";
        role = "";
        location = "";
        startDate = "";
        endDate = "";
        status = "Status";
        goto('/dashboard');
    }

    function sendFilters() {
        const params = buildParamsFromFilters({ query, company, role, location, startDate, endDate, status });
        goto(`?${params.toString()}`); 
    }
</script>

<Popover.Root>
    <Popover.Trigger asChild let:builder>
        <Button builders={[builder]} variant="outline"
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
                        bind:value={company}
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
                        bind:value={role}
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
                        bind:value={location}
                    />
                </div>
            </div>
            <div class="flex flex-col gap-2">
                <div>
                    <Label for="start-date" class="text-xs"
                        >Applied From</Label
                    >
                    <Input
                        type="date"
                        class="text-xs h-7"
                        id="start-date"
                        bind:value={startDate}
                    />
                </div>
                <div>
                    <Label for="end-date" class="text-xs"
                        >Applied Until</Label
                    >
                    <Input
                        type="date"
                        class="text-xs h-7"
                        id="end-date"
                        bind:value={endDate}
                    />
                </div>
            </div>
            <div>
                <Label for="status" class="text-xs">Status</Label>
                <Select.Root>
                    <Select.Trigger class="text-xs h-7">
                        <p
                            class="{status !== 'Status'
                                ? 'text-violet-500'
                                : 'text-gray-500'} text-xs font-medium"
                        >
                            {status}
                        </p>
                    </Select.Trigger>
                    <Select.Content>
                        <Select.Item
                            value="Applied"
                            on:click={() =>changeStatus("Applied")}
                        >
                            Applied
                        </Select.Item>
                        <Select.Item
                            value="Screen"
                            on:click={() =>changeStatus("Screen")}
                        >
                             Screen
                        </Select.Item>
                        <Select.Item
                            value="Interviewing"
                            on:click={() => changeStatus("Interviewing")}
                        >
                            Interviewing
                        </Select.Item>
                        <Select.Item
                            value="Offer"
                            on:click={() => changeStatus("Offer")}
                        >
                            Offer
                        </Select.Item>
                        <Select.Item
                            value="Rejected"
                            on:click={() => changeStatus("Rejected")}
                        >
                            Rejected
                        </Select.Item>
                        <Select.Item
                            value="Ghosted"
                            on:click={() => changeStatus("Ghosted")}
                        >
                            Ghosted
                        </Select.Item>
                    </Select.Content>
                </Select.Root>
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