<script lang="ts">
    import { onMount } from "svelte";

    import EditJob from "$lib/components/EditJob/EditJob.svelte";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";
    import { BriefcaseBusiness } from "lucide-svelte";

    import { Button } from "$lib/components/ui/button";
    import { Progress } from "$lib/components/ui/progress/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";

    import { PUBLIC_LOGO_KEY } from "$env/static/public";

    import { formatDateForDisplay } from "$lib/utils/date";

    export let objectID: string;
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;
    export let visible: boolean;

    const statusValues: Record<string, number> = {
        Rejected: 7,
        Ghosted: 26,
        Applied: 43,
        Screen: 58,
        Interviewing: 77,
        Offer: 100,
    };

    async function updateStatus(newStatus: keyof typeof statusValues) {
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

    onMount(() => {
        // inital value setting and logo fetch
        value = statusValues[status];
        fetchLogo();
    });

    let value = statusValues[status];
    let imgSrc: string | null = null;
</script>

{#if visible}
    <div class="w-full gap-2">
        <div
            class="bg-card rounded-lg border border-border shadow-sm hover:shadow-md transition-shadow m-2 dark:brightness-[0.9]"
        >
            <div class="px-4 pt-3 pb-2">
                <div class="flex items-start justify-between mb-3">
                    <div class="flex gap-3 items-center">
                        {#if imgSrc}
                            <img
                                src={imgSrc}
                                alt={company}
                                class="w-10 h-10 rounded-lg object-cover"
                            />
                        {:else}
                            <div class="w-10 h-10 p-2 rounded-lg flex items-center justify-center border border-zinc-400 border-opacity-40 dark:border-opacity-40">
                                <BriefcaseBusiness class="stroke-[1.2] text-zinc-400 opacity-70 dark:opacity-50" />
                            </div>
                        {/if}
                        <div>
                            <h3 class="font-medium">{company}</h3>
                            <div
                                class="flex items-center text-xs text-muted-foreground w-full"
                            >
                                <Map class="w-3 h-3 mr-1" />
                                {location}
                            </div>
                        </div>
                    </div>
                </div>

                <div class="mb-1 flex items-center w-full">
                    <div class="text-sm font-medium w-full truncate">
                        {#if link}
                            <a
                                href={link}
                                target="_blank"
                                rel="noopener noreferrer"
                                class="hover:underline text-primary truncate"
                                >{role}</a
                            >
                        {:else}
                            <span class="truncate">
                                {role}
                            </span>
                        {/if}
                    </div>
                    <div
                        class="flex items-center justify-end text-xs text-muted-foreground w-full"
                    >
                        <Calendar class="w-3 h-3 mr-1" />
                        {formatDateForDisplay(appliedDate)}
                    </div>
                </div>

                <div class="my-1">
                    <div class="relative">
                        <Progress
                            {value}
                            max={100}
                            class="h-0.5 absolute top-[7px]"
                        />
                        <div class="flex justify-between gap-2">
                            {#each Object.entries(statusValues) as [status, progressValue]}
                                <div
                                    class="flex flex-col items-center mt-0.5 z-10"
                                >
                                    <Button
                                        size="icon"
                                        class="w-3 h-3 z-50 rounded-full border dark:border-stone-500 {value ===
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
                                    <span class="text-[.5rem] sm:text-xs text-muted-foreground"
                                        >{status}</span
                                    >
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>

                <div class="flex justify-between items-center">
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
                                variant="ghost"
                                size="sm"
                                class="ml-2 text-red-500 hover:text-red-600 hover:bg-red-50 h-9 px-4 py-2 w-full border sm:border-0"
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
                                    permanently delete this application data
                                    from our servers.
                                </AlertDialog.Description>
                            </AlertDialog.Header>
                            <AlertDialog.Footer>
                                <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                                <AlertDialog.Action on:click={deleteApplication}
                                    >Continue</AlertDialog.Action
                                >
                            </AlertDialog.Footer>
                        </AlertDialog.Content>
                    </AlertDialog.Root>
                </div>
            </div>
        </div>
    </div>
{/if}
