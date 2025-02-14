<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Separator } from "$lib/components/ui/separator";
    import { Job } from "$lib/components/Job";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import * as Popover from "$lib/components/ui/popover";
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
    import { Label } from "$lib/components/ui/label";
    import * as Select from "$lib/components/ui/select";

    import { onMount } from "svelte";
    import { enhance } from "$app/forms";
    import type { PageData } from "./$types";
    import { goto } from "$app/navigation";

    let open = false;

    // state filters
    let query = "";
    let company = "";
    let role = "";
    let location = "";
    let startDate = "";
    let endDate = "";
    let status = "Status";

    export let data: PageData;

    // set state filters from load function on mount
    onMount(() => {
        // parses clientParams which is returned from the server load
        const params = new URLSearchParams(data.clientParams);
        query = params.get("q") || "";
        company = params.get("company") || "";
        role = params.get("role") || "";
        location = params.get("location") || "";
        startDate = params.get("startDate") || "";
        endDate = params.get("endDate") || "";
        status = params.get("status") || "Status";
    });

    function formatDateForInput(dateString: string): string {
        if (!dateString) return ""; // diff from Job.svelte (no default date)

        const parsedDate = new Date(dateString);
        if (isNaN(parsedDate.getTime()))
            return new Date().toISOString().split("T")[0];

        // adjust for timezone
        const adjustedDate = new Date(
            parsedDate.getTime() + parsedDate.getTimezoneOffset() * 60 * 1000,
        );

        return adjustedDate.toISOString().split("T")[0];
    }

    function filter() {
        // start with the current URL parameters so we merge filters
        // sorry for the huge if else block. not sure hwo to make cleaner
        const params = new URLSearchParams(window.location.search);
        if (query) {
            params.set("q", query);
        } else {
            params.delete("q");
        }
        if (company) {
            params.set("company", company);
        } else {
            params.delete("company");
        }
        if (role) {
            params.set("role", role);
        } else {
            params.delete("role");
        }
        if (location) {
            params.set("location", location);
        } else {
            params.delete("location");
        }
        if (status && status !== "Status") {
            params.set("status", status);
        } else {
            params.delete("status");
        }
        if (startDate) {
            params.set("startDate", formatDateForInput(startDate));
        } else {
            params.delete("startDate");
        }
        if (endDate) {
            params.set("endDate", formatDateForInput(endDate));
        } else {
            params.delete("endDate");
        }
        goto(`?${params.toString()}`);
    }

    function changeStatus(newStatus: string) {
        status = newStatus;
    }

    function nextPage() {
        const params = new URLSearchParams(window.location.search);
        const currentPage = parseInt(params.get("page") || "1", 10);
        const next = currentPage + 1;
        params.set("page", next.toString());
        goto(`?${params.toString()}`);
    }

    function prevPage() {
        const params = new URLSearchParams(window.location.search);
        const currentPage = parseInt(params.get("page") || "1", 10);
        const prev = currentPage > 1 ? currentPage - 1 : 1;
        params.set("page", prev.toString());
        goto(`?${params.toString()}`);
    }
</script>

<div
    class="flex flex-col justify-start gap-4 items-stretch w-full sm:w-5/6 mx-auto h-full"
>
    <div class="p-3 my-10">
        <div
            class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw]"
        >
            <div
                class="flex flex-row flex-grow gap-4 items-center w-full sm:w-auto"
            >
                <AlertDialog.Root bind:open>
                    <AlertDialog.Trigger asChild let:builder>
                        <Button
                            builders={[builder]}
                            variant="outline"
                            class="w-16"
                            on:click={() => {
                                open = true;
                            }}
                        >
                            Add
                        </Button>
                    </AlertDialog.Trigger>
                    <AlertDialog.Content>
                        <AlertDialog.Header>
                            <AlertDialog.Title
                                >Add Application</AlertDialog.Title
                            >
                            <AlertDialog.Description>
                                Add a new application to your list. Defaults to
                                'Applied' status.
                            </AlertDialog.Description>
                            <form
                                action="?/add"
                                method="POST"
                                class="flex flex-col gap-2 w-full"
                                use:enhance={(data) => {
                                    return async ({ update }) => {
                                        if (data.formElement.checkValidity()) {
                                            open = false;
                                            update();
                                        } else {
                                            // Report validity errors so the dialog remains open.
                                            data.formElement.reportValidity();
                                        }
                                    };
                                }}
                            >
                                <Input
                                    type="text"
                                    name="role"
                                    placeholder="Role"
                                    required
                                />
                                <Input
                                    type="text"
                                    name="company"
                                    placeholder="Company"
                                    required
                                />
                                <Input
                                    type="text"
                                    name="location"
                                    placeholder="Location"
                                    required
                                />
                                <Input
                                    type="text"
                                    name="link"
                                    placeholder="Link (Optional)"
                                />
                                <Input
                                    type="date"
                                    name="appliedDate"
                                    placeholder="Applied Date"
                                    required
                                />
                                <AlertDialog.Footer>
                                    <AlertDialog.Cancel
                                        on:click={() => {
                                            open = false;
                                        }}
                                    >
                                        Cancel
                                    </AlertDialog.Cancel>
                                    <AlertDialog.Action asChild>
                                        <Button type="submit">Add</Button>
                                    </AlertDialog.Action>
                                </AlertDialog.Footer>
                            </form>
                        </AlertDialog.Header>
                    </AlertDialog.Content>
                </AlertDialog.Root>

                <Separator orientation="vertical" class="h-6" />
                <Input
                    type="text"
                    placeholder="Search by company, role, or location"
                    id="query"
                    bind:value={query}
                />
                <Separator orientation="vertical" class="h-6" />
            </div>
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
                                on:click={filter}>Confirm</Button
                            >
                            <Button variant="outline"
                                on:click={() => goto("/dashboard")}
                                class="font-meidum leading-none flex-grow"
                                >Clear</Button
                            >
                        </div>
                    </div>
                </Popover.Content>
            </Popover.Root>
        </div>

        <div>
            {#each data.applications as job (job.objectID)}
                <Job
                    objectID={job.objectID}
                    company={job.company}
                    role={job.role}
                    appliedDate={job.appliedDate}
                    location={job.location}
                    status={job.status}
                    link={job.link}
                />
            {/each}
        </div>
    </div>
</div>
