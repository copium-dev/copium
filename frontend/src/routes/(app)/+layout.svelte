<script lang="ts">
    import "../../app.css";
    let { children } = $props();
    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { Separator } from "$lib/components/ui/separator";
    import { Button } from "$lib/components/ui/button";

    import { Moon } from "lucide-svelte";
    import { SunMedium } from "lucide-svelte";

    import CirclePlus from "lucide-svelte/icons/circle-plus";
    import Cloud from "lucide-svelte/icons/cloud";
    import CreditCard from "lucide-svelte/icons/credit-card";
    import Github from "lucide-svelte/icons/github";
    import Keyboard from "lucide-svelte/icons/keyboard";
    import LifeBuoy from "lucide-svelte/icons/life-buoy";
    import LogOut from "lucide-svelte/icons/log-out";
    import Mail from "lucide-svelte/icons/mail";
    import MessageSquare from "lucide-svelte/icons/message-square";
    import Plus from "lucide-svelte/icons/plus";
    import Settings from "lucide-svelte/icons/settings";
    import User from "lucide-svelte/icons/user";
    import UserPlus from "lucide-svelte/icons/user-plus";
    import Users from "lucide-svelte/icons/users";

    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";

    function signOut() {
        window.location.href = "/auth/google/logout";
    }
</script>

<ModeWatcher />
<div class="min-h-screen grid grid-rows-[auto_1fr]">
    <header class="bg-background z-50">
        <div
            class="mx-auto flex max-w-7xl items-center justify-between p-6 sm:px-8"
        >
            <div class="flex-none">
                <a href="/" class="text-2xl font-bold hover:underline">
                    jtracker
                </a>
            </div>
            <div class="flex items-center gap-4 mx-8">
                <!-- <Button variant="ghost" class="text-xs sm:text-sm"
                    >Profile</Button
                > -->
                <!-- <Separator orientation="vertical" class="h-4 sm:h-6" />
                <Button variant="ghost" class="text-xs sm:text-sm">Postings</Button> -->
            </div>
            <!-- <div class="flex gap-4 items-center">
                <Button on:click={signOut} variant="outline" class="w-16">
                    Sign out
                </Button>
                <Separator orientation="vertical" class="h-6" />
                <Button on:click={toggleMode} variant="outline" class="w-16">
                    <SunMedium class="dark:hidden" />
                    <Moon class="hidden dark:block" />
                    <span class="sr-only">Toggle theme</span>
                </Button>
            </div> -->
            <div>
                <DropdownMenu.Root>
                    <DropdownMenu.Trigger asChild let:builder>
                        <Button
                            builders={[builder]}
                            variant="outline"
                            class="w-16"
                        >
                            <Settings />
                        </Button>
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content class="w-56">
                        <!-- <DropdownMenu.Label>My Account</DropdownMenu.Label> -->

                        <DropdownMenu.Group>
                            <DropdownMenu.Item>
                                <User class="mr-2 h-4 w-4" />
                                <span>Profile</span>
                            </DropdownMenu.Item>
                        </DropdownMenu.Group>
                        <DropdownMenu.Separator />
                        <DropdownMenu.Item on:click={toggleMode}>
                            <SunMedium class="dark:hidden mr-2 h-4 w-4" />
                            <Moon class="hidden dark:block mr-2 h-4 w-4" />
                            <span class="sr-only">Toggle theme</span>
                            <p>Theme</p>
                        </DropdownMenu.Item>
                        <DropdownMenu.Item on:click={signOut}>
                            <LogOut class="mr-2 h-4 w-4" />
                            <span>Sign out</span>
                        </DropdownMenu.Item>
                    </DropdownMenu.Content>
                </DropdownMenu.Root>
            </div>
        </div>
    </header>
    <main class="bg-background overflow-auto">
        {@render children()}
    </main>
</div>
