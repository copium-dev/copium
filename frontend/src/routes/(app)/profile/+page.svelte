<script lang="ts">
    export let data;
</script>

<div class="flex flex-col justify-center gap-4 items-center h-full">
    <h1>Profile</h1>
    {#if data.email}
        <pre>{JSON.stringify(data.email, null, 2)}</pre>
    {/if}
</div>
