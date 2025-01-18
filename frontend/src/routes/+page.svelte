<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import { faGoogle } from '@fortawesome/free-brands-svg-icons';
    import Fa from 'svelte-fa';
    import { Button } from "$lib/components/ui/button";
    
    function handleGoogleSignIn() {
        window.location.href = '/api/auth/google';
    }
</script>

<div class="flex justify-center items-center h-screen">
    <Card.Root>
        <Card.Header>
        <Card.Title>Create an account</Card.Title>
            <Card.Description>Sign in with Google</Card.Description>
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