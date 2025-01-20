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
</script>

<div class="flex flex-row justify-start items-center">
    <!-- <p>Id: {id}</p> -->
    <div
        class="w-full grid grid-cols-[2fr_2fr_6fr_1fr] justify-start items-center p-3 my-3 rounded-lg border border-gray-300"
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
            <!-- <p class="border px-2 rounded border-gray-300">{status}</p> -->
            <div class="flex w-full justify-evenly gap-3 p-2">
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Rejected</p>
                </div>
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Ghosted</p>
                </div>
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Applied</p>
                </div>
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Screen</p>
                </div>
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Interview</p>
                </div>
                <div
                    class="flex flex-col justify-center items-center text-xs gap-1"
                >
                    <button class="w-5 h-5 shadow rounded"></button>
                    <p>Offer</p>
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
