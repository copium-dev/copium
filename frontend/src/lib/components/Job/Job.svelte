<script lang="ts">
    import { onMount } from "svelte";
    import { enhance } from "$app/forms";
    import { goto } from '$app/navigation'

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";

    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import { Progress } from "$lib/components/ui/progress/index.js";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";

    import placeholder from "$lib/images/placeholder.png";
    import { PUBLIC_LOGO_KEY } from "$env/static/public";

    export let objectID: string; // temporarily not used; will be used for db operations later
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp 
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

    // format to mm-dd-yyyy 
    function formatDate(timestamp: number): string {
        if (!timestamp) return "";

        const date = new Date(timestamp);
        if (isNaN(date.getTime())) return "";

        // Adjust for timezone
        const adjustedDate = new Date(
            date.getTime() + date.getTimezoneOffset() * 60 * 1000
        );

        const month = String(adjustedDate.getMonth() + 1).padStart(2, "0");
        const day = String(adjustedDate.getDate()).padStart(2, "0");
        const year = adjustedDate.getFullYear();

        return `${month}-${day}-${year}`;
    }
    
    async function updateStatus(newStatus: keyof typeof statusValues) {
        value = statusValues[newStatus];
        const formData = new FormData();
        formData.append("id", objectID);
        formData.append("status", newStatus);
        formData.append("oldStatus", status);

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

        if (!response.ok) {
            console.error("Failed to delete application");
        } else {
            console.log("Application deleted successfully");
            goto("/dashboard");
        }
    }

    onMount(() => {
        value = statusValues[status];

        const fetchLogo = async () => {
            const res = await fetch(
                `https://api.brandfetch.io/v2/search/${company}?c=${PUBLIC_LOGO_KEY}`,
            );

            if (res.ok) {
                const data = await res.json();
                imgSrc = data.length > 0 ? data[0].icon : placeholder;
            } else {
                imgSrc = placeholder;
            }
        };

        fetchLogo();
    });

    let value = statusValues[status];
    let imgSrc: string;
</script>

<Separator orientation="horizontal" class="my-2 mx-auto w-full" />
<div class="flex justify-start items-center">
    <div
        class="w-full grid grid-rows-[auto_auto_auto_auto] sm:grid-cols-[2fr_2fr_6fr_1fr] justify-center sm:justify-start items-center p-3 sm:my-2"
    >
        <div class="flex flex-row items-center">
            <img
                src={imgSrc}
                alt={`${company} logo`}
                class="w-10 h-10 rounded-lg object-cover"
            />
            <div class="flex flex-col items-baseline sm:gap-1 px-5 w-full">
                <p class="flex flex-row items-end font-bold gap-1 h-full">
                    {company}
                </p>
                <p class="flex flex-row items-end gap-1 text-xs h-full">
                    <Map class="w-[15px] h-[15px] stroke-[1.5]" />
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
                class="flex flex-row sm:flex-col items-center sm:items-baseline gap-1 px-0 sm:px-5"
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

        <div class="px-0 sm:px-5 h-full flex items-start items-center">
            <div class="flex w-full relative">
                <!-- Progress bar in background -->
                <div class="absolute w-full top-[13px]">
                    <Progress {value} max={100} class="w-full h-0.5" />
                </div>

                <!-- Buttons overlaid on top -->
                <div class="flex w-full justify-evenly gap-3 p-2 relative">
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
            <Separator
                orientation="vertical"
                class="h-12 ml-5 invisible sm:visible"
            />
        </div>

        <div
            class="mt-4 flex w-full items-stretch justify-between gap-4 sm:gap-2"
        >
            <AlertDialog.Root>
                <AlertDialog.Trigger asChild let:builder>
                    <Button
                        builders={[builder]}
                        variant="outline"
                        class="text-primary focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 px-4 py-2 text-xs flex-grow sm:border-0 sm:shadow-none sm:hover:bg-transparent sm:hover:bg-accent"
                    >
                        Edit
                    </Button>
                </AlertDialog.Trigger>
                <AlertDialog.Content>
                    <AlertDialog.Header>
                        <AlertDialog.Title>Edit Application</AlertDialog.Title>
                        <AlertDialog.Description>
                            Update your application. Unmodified fields will
                            remain unchanged.
                        </AlertDialog.Description>
                        <form
                            action="/dashboard?/editapplication"
                            method="POST"
                            class="flex flex-col gap-2 w-full"
                            use:enhance={() => {
                                return ({ update }) => {
                                    update();
                                };
                            }}
                        >
                            <!-- if any field is left empty, value will be set to the current value else overridden by the new value -->
                            <!-- hidden id field and old statuses. old status are sent for rollback purposes -->
                            <input type="hidden" name="id" value={objectID} />
                            <input type="hidden" name="oldCompany" value={company} />
                            <input type="hidden" name="oldRole" value={role} />
                            <input type="hidden" name="oldLocation" value={location} />
                            <input type="hidden" name="oldAppliedDate" value={appliedDate} />
                            <input type="hidden" name="oldLink" value={link} />
                            <div
                                class="grid grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                            >
                                <Label
                                    for="company"
                                    class="text-sm text-gray-500 font-light"
                                    >Company</Label
                                >
                                <Input
                                    type="text"
                                    name="company"
                                    placeholder="Company"
                                    value={company}
                                />
                            </div>
                            <div
                                class="grid grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                            >
                                <Label
                                    for="role"
                                    class="text-sm text-gray-500 font-light"
                                    >Role</Label
                                >
                                <Input
                                    type="text"
                                    name="role"
                                    placeholder="Role"
                                    value={role}
                                />
                        </div>
                            <div
                                class="grid grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                            >
                                <Label
                                    for="location"
                                    class="text-sm text-gray-500 font-light"
                                    >Location</Label
                                >
                                <Input
                                    type="text"
                                    name="location"
                                    placeholder="Location"
                                    value={location}
                                />
                            </div>
                            <div
                                class="grid grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                            >
                                <Label
                                    for="link"
                                    class="text-sm text-gray-500 font-light"
                                    >Link</Label
                                >
                                <Input
                                    type="text"
                                    name="link"
                                    placeholder="Link"
                                    value={link}
                                />
                            </div>
                            <div
                                class="grid grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                            >
                                <Label
                                    for="appliedDate"
                                    class="text-sm text-gray-500 font-light"
                                    >Applied Date</Label
                                >
                                <Input
                                    type="date"
                                    name="appliedDate"
                                    placeholder="Applied Date"
                                    value={appliedDate}
                                />
                            </div>
                            <AlertDialog.Footer>
                                <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                                <AlertDialog.Action type="submit"
                                    >Save</AlertDialog.Action
                                >
                            </AlertDialog.Footer>
                        </form>
                    </AlertDialog.Header>
                </AlertDialog.Content>
            </AlertDialog.Root>
            <Button
                on:click={deleteApplication}
                class="text-red-500 focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 px-4 py-2 text-xs flex-grow sm:border-0 sm:shadow-none sm:hover:bg-transparent hover:text-red-500 sm:hover:bg-accent"
            >
                Delete
            </Button>
        </div>
    </div>
</div>
