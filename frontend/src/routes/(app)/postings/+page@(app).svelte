<script lang="ts">
    import type { PageData } from "./$types";

    export let data: PageData;
</script>

<div class="flex flex-col justify-center gap-4 items-center h-full">
    <h1 class="text-2xl sm:text-4xl font-bold text-center w-3/4 sm:w-1/2">
        Postings
    </h1>
    <p>{data.email}</p>
</div>
