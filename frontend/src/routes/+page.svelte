<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import { faGoogle } from "@fortawesome/free-brands-svg-icons";
    import Fa from "svelte-fa";
    import { Button } from "$lib/components/ui/button";

    function handleGoogleSignIn() {
        window.location.href = "/auth/google/login";
    }
</script>

<div class="flex flex-col justify-center gap-4 items-center h-full">
    <h1 class="text-4xl font-bold text-center w-1/2">
        Streamline your <span class="text-violet-500">tech internship</span> hunt.
    </h1>
    <p class="text-xl text-center w-1/2">
        Track applications, discover new opportunities, and jumpstart your
        career.
    </p>
    <Card.Root class="flex flex-row mt-16">
        <Card.Header>
            <Card.Title>Created by students, for students.</Card.Title>
            <Card.Description
                >Get started; it's completely free!</Card.Description
            >
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
