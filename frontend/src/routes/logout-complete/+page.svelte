<script>
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';

    onMount(() => {
        localStorage.removeItem('authToken');
        goto('/');
    });
</script>