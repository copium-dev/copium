<script lang="ts">
    import { onMount } from "svelte";
    import { enhance } from "$app/forms";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";

    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import { Progress } from "$lib/components/ui/progress/index.js";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";

    export let id: string; // temporarily not used; will be used for db operations later
    export let company: string;
    export let role: string;
    export let appliedDate: string;
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;

    const statusValues: Record<string, number> = {
        Rejected: 10.75,
        Ghosted: 28,
        Applied: 43,
        Screen: 58,
        Interviewing: 74,
        Offer: 100,
    };

    let edit = false;

    function toggleEdit() {
        edit = !edit;
    }

    // format date for display
    function formatDate(dateString: string): string {
        if (!dateString) return '';
        
        const date = new Date(dateString);
        if (isNaN(date.getTime())) return '';
        
        // Adjust for timezone
        const adjustedDate = new Date(date.getTime() + date.getTimezoneOffset() * 60 * 1000);
        
        const month = String(adjustedDate.getMonth() + 1).padStart(2, "0");
        const day = String(adjustedDate.getDate()).padStart(2, "0");
        const year = adjustedDate.getFullYear();
        
        return `${month}-${day}-${year}`;
    }

    // format date for edit input
    function formatDateForInput(dateString: string): string {
        if (!dateString) return new Date().toISOString().split('T')[0];
        
        const parsedDate = new Date(dateString);
        if (isNaN(parsedDate.getTime())) return new Date().toISOString().split('T')[0];
        
        // Adjust for timezone
        const adjustedDate = new Date(parsedDate.getTime() + parsedDate.getTimezoneOffset() * 60 * 1000);
        
        return adjustedDate.toISOString().split('T')[0];
    }

    async function updateStatus(newStatus: keyof typeof statusValues) {
        value = statusValues[newStatus];
        const formData = new FormData();
        formData.append("id", id);
        formData.append("status", newStatus);

        const response = await fetch(`/dashboard?/editstatus`, {
            method: "POST",
            body: formData,
        });

        if (!response.ok) {
            console.error("Failed to update application");
        } else {
            console.log("Application updated successfully");
        }
    }

    async function deleteApplication() {
        const formData = new FormData();
        formData.append("id", id);

        const response = await fetch(`/dashboard?/delete`, {
            method: "POST",
            body: formData,
        });

        if (!response.ok) {
            console.error("Failed to delete application");
        } else {
            console.log("Application deleted successfully");
            window.location.reload();
        }
    }

    onMount(() => {
        value = statusValues[status];
    });

    let value = statusValues[status];
</script>

<Separator orientation="horizontal" class="my-5 mx-auto w-full" />
<div class="flex justify-start items-center">
    <div
        class="w-full grid grid-rows-[auto_auto_auto_auto] sm:grid-cols-[2fr_2fr_6fr_1fr] justify-center sm:justify-start items-center p-3 sm:my-3"
    >
        <div class="flex flex-row items-center">
            <div
                class="flex flex-row sm:flex-col items-center sm:items-baseline sm:gap-1 px-5"
            >
                <p class="flex flex-row items-end font-bold gap-1 h-full">
                    {company}
                </p>
                <p class="flex flex-row items-end gap-1 text-xs h-full">
                    <Map class="w-[15px] h-[15px] stroke-[1.5] ml-4 sm:ml-0" />
                    {location}
                </p>
            </div>
            <Separator
                orientation="vertical"
                class="h-12 ml-auto invisible sm:visible"
            />
        </div>

        <div class="flex flex-row items-center">
            <div
                class="flex flex-row sm:flex-col items-center sm:items-baseline gap-1 px-5"
            >
                <p class="flex items-end">
                    {#if link}
                        <a
                            href={link}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="hover:underline">{role}</a
                        >
                    {:else}
                        {role}
                    {/if}
                </p>
                <p class="flex flex-row items-end gap-1 text-xs h-full">
                    <Calendar
                        class="w-[15px] h-[15px] stroke-[1.5] ml-4 sm:ml-0"
                    />
                    {formatDate(appliedDate)}
                </p>
            </div>
            <Separator
                orientation="vertical"
                class="h-12 ml-auto invisible sm:visible"
            />
        </div>

        <div class="px-5 h-full flex items-center">
            <div class="flex w-full relative">
                <!-- Progress bar in background -->
                <div class="absolute w-full top-[13px]">
                    <Progress {value} max={100} class="w-full h-0.5" />
                </div>

                <!-- Buttons overlaid on top -->
                <div class="flex w-full justify-evenly gap-3 p-2 relative z-10">
                    {#each Object.entries(statusValues) as [status, progressValue]}
                        <div
                            class="flex flex-col justify-center items-center text-xs gap-1"
                        >
                            <Button
                                size="icon"
                                class="w-3 h-3 {value === progressValue
                                    ? 'bg-primary dark:bg-secondary-foreground'
                                    : 'bg-secondary dark:bg-primary-foreground'}"
                                on:click={() => {
                                    updateStatus(
                                        status as keyof typeof statusValues,
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

        {#if edit}
            <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
                <div class="relative flex flex-col gap-2 my-2 inline-block w-full overflow-visible rounded-2xl bg-white dark:bg-zinc-900 px-4 py-5 sm:p-6 text-left align-bottom shadow-xl transition-all sm:align-middle sm:max-w-3xl">
                    <h1 class="text-xl">Edit Application</h1>
                    <form
                        action="/dashboard?/editapplication"
                        method="POST"
                        class="flex flex-col gap-2 w-full"
                        use:enhance={() => {
                            toggleEdit();
                            return ({ update }) => {
                                update();
                            };
                        }}
                    >
                        <!-- if any field is left empty, value will be set to the current value else overridden by the new value -->
                        <input type="hidden" name="id" value={id} />    <!-- hidden id field for db operations -->
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="role" class="text-sm font-bold">Role</Label>
                            <Input type="text" name="role" placeholder="Role" value={role} />
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="company" class="text-sm font-bold">Company</Label>
                            <Input type="text" name="company" placeholder="Company" value={company} />
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="location" class="text-sm font-bold">Location</Label>
                            <Input type="text" name="location" placeholder="Location" value={location} />
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="link" class="text-sm font-bold">Link</Label>
                            <Input type="text" name="link" placeholder="Link" value={link} />
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="appliedDate" class="text-sm font-bold">Applied Date</Label>
                            <Input type="date" name="appliedDate" placeholder="Applied Date" value={formatDateForInput(appliedDate)} />
                        </div>
                        <div class="flex flex-row gap-2 items-stretch justify-between w-full">
                            <Button type="button" variant="outline" class="flex-grow text-red-500 hover:text-red-500" on:click={toggleEdit}>
                                Cancel
                            </Button>
                            <Button type="submit" variant="outline" class="flex-grow">
                                Save
                            </Button>
                        </div>
                    </form>
                </div>
            </div>
        {/if}

        <div class="flex ml-1.5 sm:ml-0">
            <Button
                on:click={toggleEdit}
                variant="ghost"
                class="text-xs"
            >
                Edit
            </Button>
            <Button
                variant="ghost"
                class="text-xs text-red-500 hover:text-red-500"
                on:click={deleteApplication}
            >
                Delete
            </Button>   
        </div>
    </div>
</div>
