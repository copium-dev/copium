<script lang="ts">
    export let data;
</script>

<h1>Dashboard</h1>
{#if data.email}
    <pre>{JSON.stringify(data.email, null, 2)}</pre>
{:else}
    <p>Not logged in</p>
{/if}