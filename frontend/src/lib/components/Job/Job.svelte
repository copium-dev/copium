<script lang="ts">
    export let id: string;
    export let company: string;
    export let role: string;
    export let appliedDate: Date;
    export let location: string;
    export let status = "Applied";

    import { Map } from "lucide-svelte";
    import { Calendar } from "lucide-svelte";
    import "./icons.css";

    import { Button } from "$lib/components/ui/button";
    import { Separator } from "$lib/components/ui/separator";

    // progress bar
    import { onMount } from "svelte";
    import { Progress } from "$lib/components/ui/progress/index.js";

    let value = 43;

    const statusValues = {
        Rejected: 11,
        Ghosted: 29,
        Applied: 43,
        Screen: 58,
        Interview: 76,
        Offer: 100,
    };

    function updateStatus(newStatus: keyof typeof statusValues) {
        value = statusValues[newStatus];
    }
</script>

<Separator orientation="horizontal" class="my-5" />
<div class="flex flex-row justify-start items-center">
    <!-- <p>Id: {id}</p> -->
    <div
        class="w-full grid grid-cols-[2fr_2fr_6fr_1fr] justify-start items-center p-3 my-3"
    >
        <div class="flex flex-col gap-1 border-r border-gray-300 px-5">
            <p>{company}</p>
            <p class="flex flex-row items-center gap-1 text-xs h-full">
                <Map class="lucide" />
                {location}
            </p>
        </div>

        <div class="flex flex-col gap-1 border-r border-gray-300 px-5">
            <p>{role}</p>
            <p class="flex flex-row items-center gap-1 text-xs h-full">
                <Calendar class="lucide" />
                {appliedDate.toDateString().split(" ").slice(1).join(" ")}
            </p>
        </div>

        <div class="px-5 h-full flex items-center">
            <div class="flex flex-col w-full relative">
                <!-- Progress bar in background -->
                <div class="absolute w-full top-2.5">
                    <Progress {value} max={100} class="w-full" />
                </div>

                <!-- Buttons overlaid on top -->
                <div class="flex w-full justify-evenly gap-3 p-2 relative z-10">
                    {#each Object.entries(statusValues) as [status, progressValue]}
                        <div
                            class="flex flex-col justify-center items-center text-xs gap-1"
                        >
                            <button
                                class="w-3 h-3 shadow rounded hover:ring-2 ring-offset-2 {value ===
                                progressValue
                                    ? 'bg-red-500'
                                    : 'bg-gray-200'}"
                                on:click={() => updateStatus(status as keyof typeof statusValues)}
                                aria-label={`Set status to ${status}`}
                            ></button>
                            <p>{status}</p>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <div class="flex justify-end items-center right-0">
            <button
                class=" px-2 py-1 text-xs text-gray-500 border-r border-gray-300"
            >
                Edit
            </button>
            <button class=" px-2 py-1 text-xs text-red-500"> Delete </button>
        </div>
    </div>
</div>
