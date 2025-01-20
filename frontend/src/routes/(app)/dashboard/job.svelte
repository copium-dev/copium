<script lang="ts">
    export let id: string;
    export let company: string;
    export let role: string;
    export let appliedDate: Date;
    export let location: string;
    export let status = "Applied";
</script>

<div class="flex flex-row justify-start items-center">
    <!-- <p>Id: {id}</p> -->
    <div class="w-full grid grid-cols-5 justify-start items-center p-3 m-3 rounded-lg border border-gray-300">
        <p class="border-r border-gray-300 px-5">Company: {company}</p>
        <p class="border-r border-gray-300 px-5">Role: {role}</p>
        <p class="border-r border-gray-300 px-5">Applied Date: {appliedDate.toDateString()}</p>
        <p class="border-r border-gray-300 px-5">Location: {location}</p>
        <p class="px-5">Status: {status}</p>
    </div>
</div>