<script lang="ts">
    export let id: string;
    export let company: string;
    export let role: string;
    export let appliedDate: Date;
    export let location: string;
    export let status = "Applied";
</script>

<div class="flex flex-col justify-center gap-4 items-center">
    <div class="flex flex-col justify-center gap-4 items-center p-4 rounded-lg shadow-">
        <p>Id: {id}</p>
        <p>Company: {company}</p>
        <p>Role: {role}</p>
        <p>Applied Date: {appliedDate.toDateString()}</p>
        <p>Location: {location}</p>
        <p>Status: {status}</p>
    </div>
</div>