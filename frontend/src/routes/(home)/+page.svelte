<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import { faGoogle } from "@fortawesome/free-brands-svg-icons";
    import Fa from "svelte-fa";

    import { Button } from "$lib/components/ui/button";

    function handleGoogleSignIn() {
        window.location.href = "/auth/google/login";
    }
</script>

<div class="container-wrapper grid grid-rows-2 h-5/6 sm:h-full overflow-y-hidden">
    <div class="h-full w-full">
        <div
            class="container h-full w-full flex flex-col justify-end pb-8 items-center border-x border-dashed"
        >
            <h1
                class="text-2xl sm:text-4xl font-bold text-center w-3/4 sm:w-1/2"
            >
                Streamline your <span class="text-violet-500">tech internship</span> hunt.
            </h1>
            <p class="text-base sm:text-xl text-center w-3/4 sm:w-1/2">
                Track applications, discover new opportunities, and gain data-driven insights to jumpstart your career.
            </p>
        </div>
        <div class="border-y border-dashed z-50">
            <div
                class="container flex justify-center py-8 border-x border-dashed"
            >
                <div class="w-fit">
                    <Card.Root
                        class="flex flex-col items-center sm:items-start sm:flex-row"
                    >
                        <Card.Header>
                            <Card.Title>Created by students, for students.</Card.Title>
                            <Card.Description>Get started; it's completely free!</Card.Description>
                        </Card.Header>
                        <Card.Content class="flex flex-col items-center sm:items-start gap-1">
                            <Button on:click={handleGoogleSignIn}>
                                <Fa icon={faGoogle} /> &nbsp; Sign in with Google
                            </Button>
                            <a href="/dashboard" class="text-xs tracking-wide text-muted-foreground hover:underline">Already signed in?</a>
                        </Card.Content>
                    </Card.Root>
                </div>
            </div>
        </div>
        <div class="container h-full border-x border-dashed"></div>
    </div>
</div>
