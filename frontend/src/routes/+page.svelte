<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import { faGoogle } from '@fortawesome/free-brands-svg-icons';
    import Fa from 'svelte-fa';
    import { Button } from "$lib/components/ui/button";
    
    function handleGoogleSignIn() {
        window.location.href = '/auth/google/login';
    }
</script>

<div class="flex flex-col justify-center gap-12 items-center h-screen">
    <h1 class="text-4xl font-bold text-center w-3/4">
        The #1 platform for students in tech to track 
    </h1>
    <p class="text-xl text-center">
        Track your applications, see new postings in real time, and land your dream internship.
    </p>
    <Card.Root class="flex flex-row">
        <Card.Header>
        <Card.Title>Create an account</Card.Title>
            <Card.Description>It's completely free!</Card.Description>
        </Card.Header>
        <Card.Content>
            <p>
                <Button on:click={handleGoogleSignIn}>
                    <Fa icon={faGoogle} /> &nbsp; Sign in with Google
                </Button>
            </p>
        </Card.Content>
    </Card.Root>   
</div>