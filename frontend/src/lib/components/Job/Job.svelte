<script lang="ts">
    import { onMount } from "svelte";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";

    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";
    import { Progress } from "$lib/components/ui/progress/index.js";

    export let id: string;  // temporarily not used; will be used for db operations later
    export let company: string;
    export let role: string;
    export let appliedDate: Date;
    export let location: string;
    export let status: string;

    const statusValues: Record<string, number> = {
        "Rejected": 10.75,
        "Ghosted": 28,
        "Applied": 43,
        "Screen": 58,
        "Interviewing": 74,
        "Offer": 100,
    };

    function updateStatus(newStatus: keyof typeof statusValues) {
        value = statusValues[newStatus];
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
            <div class="flex flex-row sm:flex-col items-center sm:items-baseline sm:gap-1 px-5">
                <p class="font-bold">{company}</p>
                <p class="flex flex-row items-center gap-1 text-xs h-full">
                    <Map class="w-[15px] h-[15px] stroke-[1.5] ml-4 sm:ml-0" />
                    {location}
                </p>
            </div>
            <Separator orientation="vertical" class="h-12 ml-auto invisible sm:visible" />
        </div>

        <div class="flex flex-row items-center">
            <div class="flex flex-row sm:flex-col items-center sm:items-baseline gap-1 px-5">
                <p>{role}</p>
                <p class="flex flex-row items-center gap-1 text-xs h-full">
                    <Calendar class="w-[15px] h-[15px] stroke-[1.5] ml-4 sm:ml-0" />
                    {appliedDate.toDateString().split(" ").slice(1).join(" ")}
                </p>
            </div>
            <Separator orientation="vertical" class="h-12 ml-auto invisible sm:visible" />
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
                                class="w-3 h-3 {value ===
                                progressValue
                                    ? 'bg-primary dark:bg-secondary-foreground'
                                    : 'bg-secondary dark:bg-primary-foreground'}"
                                on:click={() => updateStatus(status as keyof typeof statusValues)}
                                aria-label={`Set status to ${status}`}
                            ></Button>
                            <p>{status}</p>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <div class="flex ml-1.5 sm:ml-0">
            <Button variant="ghost" class="text-xs">Edit</Button>
            <!-- looks weird to have hover:text-red-500 but ghost automatically does hover:text-primary so this is a workaround -->
            <Button variant="ghost" class="text-xs text-red-500 hover:text-red-500">Delete</Button>
        </div>
    </div>
</div>
