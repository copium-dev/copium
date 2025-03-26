<script lang="ts">
    import { onMount } from "svelte";

    import EditJob from "$lib/components/EditJob/EditJob.svelte";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";

    import { Button, buttonVariants } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import { Progress } from "$lib/components/ui/progress/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import * as Dialog from "$lib/components/ui/dialog";
    import { toast } from "svelte-sonner";
    import { Toaster } from "$lib/components/ui/sonner/index.js";

    import { formatDateForDisplay } from "$lib/utils/date";
    import { statusUpdateStore } from "$lib/stores/statusUpdateStore"

    import { BriefcaseBusiness } from "lucide-svelte";
    import { PUBLIC_LOGO_KEY } from "$env/static/public";

    export let objectID: string;
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;
    export let visible: boolean;

    async function revertStatus() {
        const formData = new FormData();
        formData.append("id", objectID);

        // if revert returns ok, do optimistic UI from the store
        const response = await fetch(`/dashboard?/revert`, {
            method: "POST",
            body: formData,
        });

        const res = await response.json();

        if (res.type === "failure") {
            console.error("Failed to revert status");

            setTimeout(() => {
                toast.error("Could not revert status");
            }, 10);

        } else {
            console.log("Status reverted successfully");

            // optimistic UI
            status = $statusUpdateStore.prevStatus;
            value = statusValues[status];

            statusUpdateStore.set({
                ok: false,
                jobID: '',
                status: '',
                prevStatus: '',
                role: '',
                company: '',
            });
        }
    }

    const statusValues: Record<string, number> = {
        Rejected: 10.75,
        Ghosted: 26,
        Applied: 43,
        Screen: 58,
        Interviewing: 74,
        Offer: 100,
    };

    async function updateStatus(newStatus: keyof typeof statusValues) {
        // for better UX, update value before fetch
        value = statusValues[newStatus];

        const formData = new FormData();
        formData.append("id", objectID);
        formData.append("appliedDate", String(appliedDate));
        formData.append("status", newStatus);
        formData.append("oldStatus", status);

        const response = await fetch(`/dashboard?/editstatus`, {
            method: "POST",
            body: formData,
        });

        const res = await response.json();

        if (res.type === "failure") {
            console.error("Failed to update application");
        } else {
            console.log("Application updated successfully");

            // store used to display toast notification
            // '{role} at {company} status updated from {prevStatus} to {status}, click to undo'
            statusUpdateStore.set({
                ok: true,
                jobID: objectID,
                status: newStatus,
                prevStatus: status, 
                role: role,
                company: company,
            });

            // slight delay to ensure that this renders
            setTimeout(() => {
                toast.success(`${role} at ${company} status updated to ${newStatus}, click below to undo`, {
                    action: {
                        label: "Undo",
                        onClick: revertStatus,
                    },
                });
            }, 10);
        }
    }

    async function deleteApplication() {
        const formData = new FormData();
        formData.append("id", objectID);
        formData.append("company", company);
        formData.append("role", role);
        formData.append("appliedDate", String(appliedDate));
        formData.append("location", location);
        formData.append("status", status);
        formData.append("link", link || "");

        const response = await fetch(`/dashboard?/delete`, {
            method: "POST",
            body: formData,
        });

        const res = await response.json();

        if (res.type === "failure") {
            console.error("Failed to delete application");
        } else {
            console.log("Application deleted successfully");
            visible = false;
        }
    }

    async function fetchLogo(companyName: string = company) {
        try {
            const res = await fetch(
                `https://api.brandfetch.io/v2/search/${companyName}?c=${PUBLIC_LOGO_KEY}`
            );

            if (res.ok) {
                const data = await res.json();
                imgSrc = data.length > 0 ? data[0].icon : null;
            } else {
                imgSrc = null;
            }
        } catch (error) {
            console.error("Error fetching logo:", error);
            imgSrc = null;
        }
    }

    function handleJobUpdate(updatedJob: {
        company: string;
        role: string;
        location: string;
        link: string | undefined | null;
        appliedDate: number;
        status: string;
    }) {
        console.log("Job update received:", updatedJob);

        // if company changed then we refetch logo
        const companyChanged = company !== updatedJob.company;

        company = updatedJob.company;
        role = updatedJob.role;
        location = updatedJob.location;
        link = updatedJob.link;
        appliedDate = updatedJob.appliedDate;
        status = updatedJob.status;
        value = statusValues[status];

        if (companyChanged) {
            console.log("Company changed, fetching new logo");
            fetchLogo(updatedJob.company);
        }
    }

    async function showTimeline() {
        const formData = new FormData();
        formData.append("id", objectID);

        const response = await fetch(`/dashboard?/timeline`, { 
            method: "POST",
            body: formData,
        });

        const res = await response.json();

        if (res.type === "failure") {
            console.error("Failed to get application timeline");
        } else {
            console.log("Timeline fetched successfully:", res.data);
        }
    }

    onMount(() => {
        // inital value setting and logo fetch
        value = statusValues[status];
        fetchLogo();
    });

    let value = statusValues[status];
    let imgSrc: string | null = null;
    let timeline: any[] = []
</script>

{#if visible}
    <Toaster />
    <Separator
        orientation="horizontal"
        class="mx-auto w-full border-t border-dashed bg-transparent"
    />
    <div class="px-8 flex flex-col justify-start items-center py-4 sm:py-0">
        <div
            class="gap-4 sm:gap-0 w-full grid grid-rows-[auto_auto_auto_auto] sm:grid-cols-[2fr_2fr_6fr_1fr] justify-center sm:justify-start items-center dark:brightness-[0.9]"
        >
            <div
                class="sm:h-20 border-none sm:border-r sm:border-dashed flex flex-col sm:flex-row w-64"
            >
                <div class="flex flex-row items-center">
                    {#if imgSrc}
                        <img
                            src={imgSrc}
                            alt={company}
                            class="w-10 h-10 rounded-lg object-cover"
                        />
                    {:else}
                        <div class="w-10 h-10 p-2 rounded-lg flex items-center justify-center border border-zinc-400 border-opacity-50 dark:border-opacity-40">
                            <BriefcaseBusiness class="stroke-[1.5] text-zinc-400 opacity-70 dark:opacity-50" />
                        </div>
                    {/if}
                    <div
                        class="flex flex-col items-baseline sm:gap-1 px-5 w-full truncate"
                    >
                        <p
                            class="flex flex-row items-end text-md font-bold gap-1 h-full w-full truncate"
                        >
                            <span class="truncate overflow-hidden"
                                >{company}</span
                            >
                        </p>
                        <p
                            class="flex flex-row items-end text-xs gap-1 h-full w-full truncate"
                        >
                            <Map
                                class="w-[15px] h-[15px] stroke-[1.5] flex-shrink-0"
                            />
                            <span class="truncate overflow-hidden"
                                >{location}</span
                            >
                        </p>
                    </div>
                    <div class="sm:hidden flex flex-row items-center gap-4">
                        <EditJob
                            {objectID}
                            {company}
                            {role}
                            {appliedDate}
                            {location}
                            {status}
                            {link}
                            onUpdateSuccess={handleJobUpdate}
                            onDeleteSuccess={() => (visible = false)}
                        />
                    </div>
                </div>
            </div>

            <div
                class="sm:h-20 border-none sm:border-r sm:border-dashed flex flex-col sm:flex-row w-full"
            >
                <div class="flex flex-row items-center w-full">
                    <div
                        class="flex flex-row sm:flex-col items-center sm:items-baseline gap-1 px-0 sm:px-5 w-[384px] sm:w-[300px]"
                    >
                        <div class="flex flex-row w-full justify-stretch">
                            <p class="flex items-end w-full truncate">
                                {#if link}
                                    <a
                                        href={link}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        class="hover:underline truncate text-md font-medium"
                                        >{role}</a
                                    >
                                {:else}
                                    <span class="truncate text-md font-medium"
                                        >{role}</span
                                    >
                                {/if}
                            </p>
                            <Dialog.Root>
                                <Dialog.Trigger
                                    class={buttonVariants({ variant: "ghost" }) + "text-xs sm:text-sm"}
                                    on:click={showTimeline}
                                >
                                    Timeline
                                </Dialog.Trigger>
                                <Dialog.Content>
                                    <Dialog.Title>Timeline</Dialog.Title>
                                    <Dialog.Description>
                                        {#if timeline.length > 0}
                                            <p>{timeline}</p>
                                        {:else}
                                            <p>No timeline data available</p>
                                        {/if}
                                    </Dialog.Description>
                                    <Dialog.Close>Close</Dialog.Close>
                                </Dialog.Content>
                            </Dialog.Root>
                        </div>
                        <p
                            class="flex flex-row items-end text-xs gap-1 h-full w-full"
                        >
                            <Calendar class="w-[15px] h-[15px] stroke-[1.5]" />
                            {formatDateForDisplay(appliedDate)}
                        </p>
                    </div>
                </div>
            </div>

            <div
                class="sm:h-20 border-none sm:border-r sm:border-dashed flex flex-col sm:flex-row items-center w-full"
            >
                <div
                    class="px-0 sm:px-5 h-full flex items-start items-center w-full"
                >
                    <div class="flex w-full relative">
                        <div class="absolute w-full top-[13px]">
                            <Progress {value} max={100} class="w-full h-0.5" />
                        </div>

                        <div
                            class="flex w-full justify-evenly gap-3 p-2 relative"
                        >
                            {#each Object.entries(statusValues) as [status, progressValue]}
                                <div
                                    class="flex flex-col justify-center items-center text-xs gap-1"
                                >
                                    <Button
                                        size="icon"
                                        class="w-3 h-3 border dark:border-stone-500 {value ===
                                        progressValue
                                            ? 'bg-primary dark:bg-secondary-foreground'
                                            : 'bg-secondary dark:bg-primary-foreground'}"
                                        on:click={() => {
                                            updateStatus(
                                                status as keyof typeof statusValues
                                            );
                                        }}
                                        aria-label={`Set status to ${status}`}
                                    ></Button>
                                    <p>{status}</p>
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="ml-4 flex w-full items-stretch justify-between gap-4 sm:gap-2 hidden sm:flex"
            >
                <EditJob
                    {objectID}
                    {company}
                    {role}
                    {appliedDate}
                    {location}
                    {status}
                    {link}
                    onUpdateSuccess={handleJobUpdate}
                    onDeleteSuccess={() => (visible = false)}
                />

                <AlertDialog.Root>
                    <AlertDialog.Trigger asChild let:builder>
                        <Button
                            builders={[builder]}
                            class="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-300 dark:hover:text-red-700 focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background border shadow-sm h-9 px-4 py-2 text-xs flex-grow sm:border-0 sm:shadow-none"
                        >
                            Delete
                        </Button>
                    </AlertDialog.Trigger>
                    <AlertDialog.Content>
                        <AlertDialog.Header>
                            <AlertDialog.Title
                                >Are you absolutely sure?</AlertDialog.Title
                            >
                            <AlertDialog.Description>
                                This action cannot be undone. This will
                                permanently delete this application data from
                                our servers.
                            </AlertDialog.Description>
                        </AlertDialog.Header>
                        <AlertDialog.Footer>
                            <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                            <AlertDialog.Action on:click={deleteApplication}>
                                Continue
                            </AlertDialog.Action>
                        </AlertDialog.Footer>
                    </AlertDialog.Content>
                </AlertDialog.Root>
            </div>
        </div>
    </div>
{/if}
