<script lang="ts">
    import * as Tabs from "$lib/components/ui/tabs/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
</script>

<div class="container h-full border-x border-dashed py-4 flex justify-between">
    <Tabs.Root value="underConsideration" class="w-full dark:brightness-[0.9]">
        <Tabs.List class="grid w-full grid-cols-3 mt-6">
            <Tabs.Trigger value="underConsideration"
                >Under Consideration</Tabs.Trigger
            >
            <Tabs.Trigger value="comingSoon">Coming Soon</Tabs.Trigger>
            <Tabs.Trigger value="launched">Launched</Tabs.Trigger>
        </Tabs.List>

        <Tabs.Content value="underConsideration">
            <div class="grid grid-cols-2 gap-4 my-6">
                <Card.Root class="flex flex-col justify-between">
                    <Card.Header>
                        <Card.Title
                            >Parse Application Data from csv / to csv</Card.Title
                        >
                        <Card.Description>
                            Import your application data from Google Sheets or
                            Excel to copium.dev
                        </Card.Description>

                        <Card.Content class="p-0 py-6">
                            We are considering adding a feature that allows you
                            to import your application data from Google Sheets
                            or Excel to copium.dev. This would make it easier to
                            transition from your previous applicaiton tracking
                            tool to copium.dev.
                        </Card.Content>
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
                <Card.Root
                    class="h-full w-full gap-4 flex flex-col justify-between"
                >
                    <Card.Header>
                        <Card.Title>Multi-select Applicaiton Delete</Card.Title>
                        <Card.Description>
                            Delete multiple applications at once
                        </Card.Description>

                        <Card.Content class="p-0 py-6">
                            We are considering adding a feature that allows you
                            to delete multiple applications at once. This would
                            make it easier to clean up your application list.
                        </Card.Content>
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
                <Card.Root
                    class="h-full w-full gap-4 flex flex-col justify-between"
                >
                    <Card.Header>
                        <Card.Title>Revert One Application Edit</Card.Title>
                        <Card.Description>
                            Revert one application edit action, in case you make
                            a mistake or change your mind
                        </Card.Description>

                        <Card.Content class="p-0 py-6">
                            We are considering adding a feature that allows you
                            to revert one application edit action. This would
                            make it easier to undo a mistake or changed-my-mind,
                            without disturbing the analytics of your
                            applications.
                        </Card.Content>
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
            </div>
        </Tabs.Content>

        <Tabs.Content value="comingSoon">
            <div class="grid grid-cols-2 gap-4 my-6">
                <Card.Root class="flex flex-col justify-between">
                    <Card.Header>
                        <Card.Title>
                            Lots of bug fixes and performance improvements
                        </Card.Title>
                        <Card.Description>
                            We are constantly working on fixing bugs as well as making copium.dev faster and more reliable
                        </Card.Description>

                        <Card.Content class="p-0 py-6">
                            We are constantly working on fixing bugs as well as
                            making copium.dev faster and more reliable. We are
                            also working on adding new features and improving the
                            user experience, so stay tuned for updates!
                        </Card.Content>
                    </Card.Header>
                    <Card.Footer class="flex flex-col items-start gap-2">
                        <p class="text-muted-foreground text-sm">
                            Expected Release: ASAP
                        </p>
                    </Card.Footer>
                </Card.Root>
            </div>
        </Tabs.Content>

        <Tabs.Content value="launched">
            <div class="grid grid-cols-2 gap-4 my-6">
                <Card.Root class="flex flex-col justify-between">
                    <Card.Header>
                        <Card.Title>copium.dev is now available!</Card.Title>
                        <Card.Description>
                            copium.dev is now available for anyone to use! It
                            was released on March 20, 2025.
                        </Card.Description>

                        <Card.Content class="p-0 py-6">
                            We are excited to announce that copium.dev is now
                            available for anyone to use! It was released on
                            March 20, 2025. We are still working on adding more
                            features and improving the user experience, so stay
                            tuned for updates!
                        </Card.Content>
                    </Card.Header>
                    <Card.Footer class="flex flex-col items-start gap-2">
                        <p class="text-muted-foreground text-sm">
                            Released March 20, 2025
                        </p>
                    </Card.Footer>
                </Card.Root>
            </div>
        </Tabs.Content>
    </Tabs.Root>
</div>
