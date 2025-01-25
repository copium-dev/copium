<script lang="ts">
    import '../../app.css';
    let { children } = $props();
    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { Button } from "$lib/components/ui/button";

    import { Moon } from 'lucide-svelte';
    import { SunMedium } from 'lucide-svelte';
</script>

<ModeWatcher />
<div class="min-h-screen grid grid-rows-[auto_1fr]">
    <header class="bg-background z-50">
        <div class="mx-auto flex max-w-7xl items-center justify-between p-6 lg:px-8">
            <div class="flex-none">
                <a href="/" aria-label="Home" class="group hover:opacity-80 transition-all">
                    <svg version="1.0" xmlns="http://www.w3.org/2000/svg" class="w-12 sm:w-16 h-auto" width="500.000000pt" height="200.000000pt" viewBox="0 150 1.000000 200.000000" preserveAspectRatio="xMidYMid meet">
                        <g transform="translate(-250.000000,500.000000) scale(0.100000,-0.10000)" stroke="none">
                            <path d="M2637 3416 c-10 -8 -90 -110 -177 -227 -148 -196 -160 -216 -160 -255 0 -35 5 -47 31 -68 59 -50 101 -30 189 90 l65 87 3 -152 3 -152 97 3 97 3 -3 148 c-1 81 1 147 5 146 5 0 39 -41 76 -90 89 -117 133 -134 194 -73 21 21 25 32 21 67 -4 36 -30 75 -162 249 -86 114 -166 214 -178 223 -27 18 -77 19 -101 1z" fill="#8B5DF5"/>
                            <path d="M327 2659 c-114 -17 -205 -89 -249 -198 -20 -49 -23 -74 -23 -196 0 -134 1 -143 29 -203 56 -122 152 -177 306 -176 117 1 183 27 247 98 47 51 83 127 83 174 0 21 -3 22 -93 22 l-94 0 -12 -39 c-26 -87 -125 -120 -208 -70 -46 29 -66 90 -67 202 -1 111 23 174 76 207 39 24 107 26 147 5 26 -13 61 -67 61 -94 0 -7 34 -11 100 -11 l99 0 -5 33 c-33 178 -191 276 -397 246z" fill="currentColor"/>
                            <path d="M1128 2659 c-115 -17 -217 -93 -267 -198 -24 -52 -26 -66 -26 -186 0 -114 3 -136 23 -179 52 -113 146 -184 274 -206 182 -32 357 59 419 218 30 77 33 241 6 320 -24 66 -85 145 -143 182 -70 44 -186 64 -286 49z m170 -179 c54 -33 77 -96 77 -210 -1 -78 -5 -102 -23 -137 -27 -51 -59 -74 -115 -83 -115 -17 -186 38 -207 158 -20 120 10 223 80 269 45 31 140 32 188 3z" fill="currentColor"/>
                            <path d="M2083 2659 c-60 -11 -96 -33 -137 -81 l-36 -42 0 57 0 57 -90 0 -90 0 0 -515 0 -515 95 0 95 0 0 187 0 187 34 -37 c47 -52 96 -71 183 -70 57 0 82 5 127 27 66 32 99 69 140 156 30 63 31 69 31 210 0 130 -2 150 -23 194 -64 140 -191 212 -329 185z m91 -189 c49 -43 68 -99 68 -195 0 -96 -19 -152 -68 -195 -30 -26 -43 -30 -92 -30 -71 0 -113 24 -144 82 -19 36 -23 58 -23 138 0 162 49 230 167 230 49 0 62 -4 92 -30z" fill="currentColor"/>
                            <path d="M4181 2659 c-60 -12 -98 -39 -146 -104 -13 -18 -14 -15 -15 38 l0 57 -90 0 -90 0 0 -370 0 -370 95 0 94 0 3 266 3 266 33 29 c27 24 41 29 83 29 113 0 119 -18 119 -340 l0 -250 95 0 95 0 0 228 c0 125 5 245 10 267 14 60 58 95 120 95 61 0 97 -28 110 -85 5 -22 10 -145 10 -272 l0 -233 95 0 95 0 0 273 c0 314 -6 349 -67 409 -51 49 -103 70 -182 71 -78 1 -133 -23 -178 -77 l-29 -35 -27 34 c-51 63 -144 93 -236 74z" fill="currentColor"/>
                            <path d="M2590 2280 l0 -370 100 0 100 0 0 370 0 370 -100 0 -100 0 0 -370z" fill="currentColor"/>
                            <path d="M2982 2348 c3 -293 4 -304 26 -346 57 -106 203 -148 332 -95 19 7 54 33 78 56 l42 41 0 -47 0 -47 90 0 90 0 0 370 0 370 -95 0 -95 0 0 -240 c0 -259 -4 -283 -53 -329 -50 -46 -148 -44 -196 3 l-26 26 -3 270 -3 270 -95 0 -95 0 3 -302z" fill="currentColor"/>
                        </g>
                    </svg>
                </a>
            </div>
            <div class="flex items-center gap-2 sm:gap-4 mx-4 sm:mx-8">
                <!-- <Button variant="ghost" class="text-xs sm:text-sm">Profile</Button> -->
                <!-- <Separator orientation="vertical" class="h-4 sm:h-6" /> -->
                <!-- <Button variant="ghost" class="text-xs sm:text-sm">Postings</Button> -->
            </div>
            <div class="flex-none">
                <Button on:click={toggleMode} variant="outline" class="w-12 sm:w-16" aria-label="Toggle theme">
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