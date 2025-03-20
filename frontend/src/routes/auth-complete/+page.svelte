<script>
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    
    onMount(() => {
        const token = new URLSearchParams(window.location.search).get('token');
        if (token) {
            localStorage.setItem('authToken', token);
            // set a cookie to use within this same domain
            // aka for server-side functions. this makes it a lot easier to use JWT token
            // in backend calls
            document.cookie = `authToken=${token}; path=/; max-age=${30*24*60*60}; SameSite=Lax`;
            goto('/dashboard');
        }
    });
</script>