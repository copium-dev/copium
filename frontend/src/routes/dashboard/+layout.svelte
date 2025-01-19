<script lang="ts">
    let { children } = $props();
    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { Separator } from "$lib/components/ui/separator";
    import { Button } from "$lib/components/ui/button";

    import { Moon } from 'lucide-svelte';
    import { SunMedium } from 'lucide-svelte';

    function signOut () {
        window.location.href = "/auth/google/logout";
    }
</script>

<ModeWatcher />
<div class="min-h-screen grid grid-rows-[auto_1fr]">
    <header class="bg-background z-50">
        <div class="mx-auto flex max-w-7xl items-center justify-between p-6 lg:px-8">
            <div class="flex-none">
                <a href="/" class="text-2xl font-bold hover:underline">
                    jtracker
                </a>
            </div>
            <div class="flex items-center gap-4 mx-8">
                <Button variant="ghost">Profile</Button>
                <Separator orientation="vertical" class="h-6" />
                <Button variant="ghost">Postings</Button>
            </div>
            <div class="flex gap-4 self-items-center">
                <Button on:click={signOut} variant="outline" class="w-[4rem]">
                    Sign out
                </Button>
                <Button on:click={toggleMode} variant="outline" class="w-[4rem]">
                    <SunMedium class="dark:hidden"/>
                    <Moon class="hidden dark:block"/>
                    <span class="sr-only">Toggle theme</span>
                </Button>
            </div>
        </div>
    </header>
    <main class="bg-background overflow-auto">
        {@render children()}
    </main>
</div>