<script lang="ts">
    import { onMount } from "svelte";

    import EditJob from "$lib/components/EditJob/EditJob.svelte";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";

    import { Button } from "$lib/components/ui/button";
    import { Progress } from "$lib/components/ui/progress/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";

    import placeholder from "$lib/images/placeholder.png";
    import { PUBLIC_LOGO_KEY } from "$env/static/public";

    export let objectID: string;
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;
    export let visible: boolean;

    const statusValues: Record<string, number> = {
        Rejected: 9,
        Ghosted: 26,
        Applied: 43,
        Screen: 58,
        Interviewing: 77,
        Offer: 100,
    };

    // format to mm-dd-yyyy
    function formatDate(timestamp: number): string {
        if (!timestamp) return "";

        const date = new Date(timestamp * 1000);
        if (isNaN(date.getTime())) return "";

        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");
        const year = date.getFullYear();

        return `${month}-${day}-${year}`;
    }

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
                imgSrc = data.length > 0 ? data[0].icon : placeholder;
            } else {
                imgSrc = placeholder;
            }
        } catch (error) {
            console.error("Error fetching logo:", error);
            imgSrc = placeholder;
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
    let imgSrc: string;
</script>

{#if visible}
    <div class="bg-card rounded-lg border border-border shadow-sm hover:shadow-md transition-shadow">
        <div class="px-6 pt-4 pb-3">
            <div class="flex items-start justify-between mb-2">
                <div class="flex gap-3 items-center">
                    <img
                        src={imgSrc}
                        alt={company}
                        class="w-8 h-8 rounded-lg object-cover"
                    />
                    <div>
                        <h3 class="font-medium">{company}</h3>
                        <div class="flex items-center text-xs text-muted-foreground">
                            <Map class="w-3 h-3 mr-1" />
                            {location}
                        </div>
                    </div>
                </div>
            </div>
        
            <div class="mb-1 flex gap-2 items-center">
                <div class="text-sm font-medium">
                    {#if link}
                        <a
                            href={link}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="hover:underline text-primary"
                        >{role}</a>
                    {:else}
                        {role}
                    {/if}
                </div>
                <div class="flex items-center text-xs text-muted-foreground">
                    <Calendar class="w-3 h-3 mr-1" />
                    {formatDate(appliedDate)}
                </div>
            </div>
            
            <div class="my-1">
                <div class="relative">
                    <Progress value={value} max={100} class="h-0.5 absolute top-[7px]" />
                    <div class="flex justify-between">
                        {#each Object.entries(statusValues) as [status, progressValue]}
                            <div class="flex flex-col items-center mt-0.5 z-10">
                                <Button
                                    size="icon"
                                    class="w-3 h-3 z-50 {value === progressValue
                                        ? 'bg-primary dark:bg-secondary-foreground'
                                        : 'bg-secondary dark:bg-primary-foreground'}"
                                    on:click={() => {
                                        updateStatus(
                                            status as keyof typeof statusValues
                                        );
                                    }}
                                    aria-label={`Set status to ${status}`}
                                ></Button>
                                <span class="text-[10px] mt-1 hidden sm:inline">{status}</span>
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
                    onDeleteSuccess={() => visible = false}
                />
                
                <AlertDialog.Root>
                    <AlertDialog.Trigger asChild let:builder>
                        <Button
                            builders={[builder]}
                            variant="ghost"
                            size="sm"
                            class="text-red-500 hover:text-red-600 hover:bg-red-50 p-2 h-8 w-1/2"
                        >
                            Delete
                        </Button>
                    </AlertDialog.Trigger>
                    <AlertDialog.Content>
                        <!-- Keep your existing Alert Dialog content -->
                        <AlertDialog.Header>
                            <AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
                            <AlertDialog.Description>
                                This action cannot be undone. This will permanently delete this application data from our servers.
                            </AlertDialog.Description>
                        </AlertDialog.Header>
                        <AlertDialog.Footer>
                            <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                            <AlertDialog.Action on:click={deleteApplication}>Continue</AlertDialog.Action>
                        </AlertDialog.Footer>
                    </AlertDialog.Content>
                </AlertDialog.Root>
            </div>
        </div>
    </div>
{/if}