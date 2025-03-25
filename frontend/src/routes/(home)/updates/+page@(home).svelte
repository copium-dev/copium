<script lang="ts">
    import * as Tabs from "$lib/components/ui/tabs/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";

    const underConsideration = [
        {
            title: "Parse application data from csv",
            description: "Import your application data from Google Sheets or Excel to copium.dev",
            content: "We are considering adding a feature that allows you to import your application data from Google Sheets or \
            Excel to copium.dev, making it easier to transition to copium.dev.",
        },
        {
            title: "Multi-select application delete",
            description: "Delete multiple applications at once",
            content: "We are considering adding a feature that allows you to delete multiple applications at once, making it \
            easier to clean up your application list.",
        },
        {
            title: "New grad job board",
            description: "Discover new grad opportunities",
            content: "We are considering adding a new grad job board to copium.dev to support a larger audience of tech students. \
            This is a highly requested feature and could be easily implemented, but price is a concern because of the extra data \
            storage and processing required."
        }
    ];

    const comingSoon = [
        {
            title: "Lots of bug fixes and performance improvements",
            description: "We are constantly working on fixing bugs as well as making copium.dev faster and more reliable",
            content: "We are constantly working on fixing bugs as well as making copium.dev faster and more reliable. \
            We are also working on adding new features and improving the user experience, so stay tuned for updates!",
            expectedRelease: "ASAP"
        },
        {
            title: "Full application timeline history",
            description: "View the entire history of an application",
            content: "We are working on adding a feature that allows you to view the entire history of an application, including \
            all status changes and edits. This would complement a recent feature that allows reverting one application edit, \
            allowing you to have deeper control on the analytics captured by copium.dev.",
            expectedRelease: "April 1, 2025"
        }
    ];

    const launched = [
        {
            title: "copium.dev is now available!",
            description: "copium.dev is now available for anyone to use! It was released on March 20, 2025.",
            content: "We are excited to announce that copium.dev is now available for anyone to use! It was released on March 20, 2025. We are still working on adding more features and improving the user experience, so stay tuned for updates!",
            releaseDate: "March 20, 2025"
        },
        {
            title: "Revert one application edit",
            description: "Revert one application status edit, in case you made a mistake.",
            content: "Manually reverting an application status edit by simply clicking the previous status would oftentimes disturb analytics. To address this, an actual revert feature has been added. A more concrete fix is in the works, to allow reverting any edit in the application's timeline.",
            releaseDate: "March 24, 2025"
        },
    ]
</script>

<div class="container h-full border-x border-dashed py-4 flex justify-between">
    <Tabs.Root value="underConsideration" class="w-full dark:brightness-[0.9]">
        <Tabs.List class="grid w-full grid-cols-3 mt-6">
            <Tabs.Trigger value="underConsideration">
                <span class="w-full truncate">Under Consideration</span>
            </Tabs.Trigger
            >
            <Tabs.Trigger value="comingSoon">
                <span class="w-full truncate">Coming Soon</span>
            </Tabs.Trigger>
            <Tabs.Trigger value="launched">
                <span class="w-full truncate">Launched</span>
            </Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="underConsideration">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 my-6">
                {#each underConsideration as { title, description, content }}
                    <Card.Root class="flex flex-col justify-between">
                        <Card.Header>
                            <Card.Title>{title}</Card.Title>
                            <Card.Description>{description}</Card.Description>
                            <Card.Content class="p-0 py-6">{content}</Card.Content>
                        </Card.Header>
                        <!-- <Card.Footer class="flex flex-col items-start gap-2">
                            <p class="text-muted-foreground text-sm">
                                How important is this to you?
                            </p>
                            <div class="flex gap-2 w-full">
                                <Button
                                    class="w-1/4 hover:bg-gray-500 hover:text-white"
                                    >NOT IMPORTANT</Button
                                >
                                <Button
                                    class="w-1/4 hover:bg-green-500 hover:text-white"
                                    >NICE-TO-HAVE</Button
                                >
                                <Button
                                    class="w-1/4 hover:bg-blue-500 hover:text-white"
                                    >IMPORTANT</Button
                                >
                                <Button
                                    class="w-1/4 hover:bg-red-500 hover:text-white"
                                    >CRITICAL</Button
                                >
                            </div>
                        </Card.Footer> -->
                    </Card.Root>
                {/each}
            </div>
        </Tabs.Content>

        <Tabs.Content value="comingSoon">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 my-6">
                {#each comingSoon as { title, description, content, expectedRelease }}
                    <Card.Root class="flex flex-col justify-between">
                        <Card.Header>
                            <Card.Title>{title}</Card.Title>
                            <Card.Description>{description}</Card.Description>
                            <Card.Content class="p-0 py-6">{content}</Card.Content>
                        </Card.Header>
                        <Card.Footer class="flex flex-col items-start gap-2">
                            <p class="text-muted-foreground text-sm">
                                Expected Release: {expectedRelease}
                            </p>
                        </Card.Footer>
                    </Card.Root>
                {/each}
            </div>
        </Tabs.Content>

        <Tabs.Content value="launched">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 my-6">
                {#each launched as { title, description, content, releaseDate }}
                    <Card.Root class="flex flex-col justify-between">
                        <Card.Header>
                            <Card.Title>{title}</Card.Title>
                            <Card.Description>{description}</Card.Description>
                            <Card.Content class="p-0 py-6">{content}</Card.Content>
                        </Card.Header>
                        <Card.Footer class="flex flex-col items-start gap-2">
                            <p class="text-muted-foreground text-sm">
                                Released {releaseDate}
                            </p>
                        </Card.Footer>
                    </Card.Root>
                {/each}
            </div>
        </Tabs.Content>
    </Tabs.Root>
</div>
