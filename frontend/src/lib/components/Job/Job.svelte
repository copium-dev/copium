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
</script>

<div class="flex flex-row justify-start items-center">
    <!-- <p>Id: {id}</p> -->
    <div
        class="w-full grid grid-cols-5 justify-start items-center p-3 m-3 rounded-lg border border-gray-300"
    >
        <div class="flex flex-col gap-1 border-r border-gray-300 px-5">
            <p>{company}</p>
            <p class="flex flex-row items-center gap-1 text-sm h-full">
                <Map class="lucide" />
                {location}
            </p>
        </div>

        <div class="flex flex-col gap-1 border-r border-gray-300 px-5">
            <p>{role}</p>
            <p class="flex flex-row items-center gap-1 text-sm h-full">
                <Calendar class="lucide" />
                {appliedDate.toDateString()}
            </p>
        </div>

        <p class="px-5 h-full flex items-center">Status: {status}</p>
    </div>
</div>
