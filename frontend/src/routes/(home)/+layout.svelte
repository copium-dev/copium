<script lang="ts">
    import '../../app.css';
    let { children } = $props();
    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { Separator } from "$lib/components/ui/separator";
    import { Button } from "$lib/components/ui/button";

    import { Moon } from 'lucide-svelte';
    import { SunMedium } from 'lucide-svelte';
</script>

<ModeWatcher />
<div class="min-h-screen grid grid-rows-[auto_1fr]">
    <header class="bg-background z-50">
        <div class="mx-auto flex max-w-7xl items-center justify-between p-6 lg:px-8">
            <div class="flex-none">
                <a href="/" class="text-lg sm:text-2xl font-bold hover:underline">
                    jtracker
                </a>
            </div>
            <div class="flex items-center gap-2 sm:gap-4 mx-4 sm:mx-8">
                <Button variant="ghost" class="text-xs sm:text-sm">Profile</Button>
                <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" class="text-xs sm:text-sm">Postings</Button>
            </div>
            <div class="flex-none">
                <Button on:click={toggleMode} variant="outline" class="w-12 sm:w-16">
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