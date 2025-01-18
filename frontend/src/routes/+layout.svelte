<script lang="ts">
    import '../app.css';
    let { children } = $props();
    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { Separator } from "$lib/components/ui/separator";
    import { Button } from "$lib/components/ui/button";
</script>

<ModeWatcher />
<header class="fixed top-0 left-0 right-0 bg-background z-50">
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
        <div class="flex-none">
            <Button on:click={toggleMode} variant="outline" class="w-[4rem]">
                <p class="dark:hidden">Light</p>
                <p class="hidden dark:block">Dark</p>
                <span class="sr-only">Toggle theme</span>
            </Button>
        </div>
    </div>
</header>
<main class="bg-background h-screen flex justify-center items-center">
    {@render children()}
</main>