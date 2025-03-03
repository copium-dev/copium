<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import * as Card from "$lib/components/ui/card/index.js";

    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import { enhance } from "$app/forms";   
    import type { PageData } from "./$types";

    function signOut() {
        window.location.href = "/auth/google/logout";
    }

    // method not allowed error with method DELETE
    // even though cors is set up correctly so just post with empty body
    async function deleteAccount() {
        const formData = new FormData();
        const response = await fetch("?/delete", {
            method: "POST",
            body: formData
        });

        if (!response.ok) {
            console.error("Failed to delete account");
        }
        
        signOut();
    }

    export let data: PageData;
</script>

<ModeWatcher />
<div
    class="flex flex-col justify-start gap-4 items-stretch w-full mx-auto h-full"
>
    <div class="my-10">
        <div
            class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-start w-full"
        >
            <!-- todo: render ....
                - total number of applications
                - delete account button
                - sign out button
                - dark mode toggle -->
            <Card.Root class="w-full sm:w-1/3">
                <Card.Header>
                    <Card.Title>{data.email}</Card.Title>
                    <Card.Description>Total Applications: {data.applicationsCount}</Card.Description>
                </Card.Header>
                <Card.Content>
                    <div class="grid grid-cols-2 gap-2">
                        <div class="col-span-1 flex justify-center">
                            <Button variant="outline" class="w-full" on:click={signOut}>
                                Sign out
                            </Button>
                        </div>
                        <div class="col-span-1 flex justify-center">
                            <Button variant="outline" class="w-full" on:click={toggleMode}>
                                Toggle theme
                            </Button>
                        </div>
                    </div>
                    <!-- Delete Account as the last full-width row with rotation effect -->
                    <div class="grid grid-cols-1 mt-2">
                        <AlertDialog.Root>
                            <AlertDialog.Trigger asChild let:builder>
                                <Button builders={[builder]} variant="outline" class="text-red-500 hover:text-red-500">
                                    Delete account
                                </Button>
                              </AlertDialog.Trigger>
                              <AlertDialog.Content>
                                <AlertDialog.Header>
                                  <AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
                                  <AlertDialog.Description>
                                    This action cannot be undone. This will permanently delete your account
                                    and remove your data from our servers.
                                  </AlertDialog.Description>
                                </AlertDialog.Header>
                                <AlertDialog.Footer>
                                  <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                                  <AlertDialog.Action
                                    on:click={deleteAccount}
                                  >
                                    Continue
                                    </AlertDialog.Action>
                                </AlertDialog.Footer>
                              </AlertDialog.Content>
                        </AlertDialog.Root>
                    </div>
                </Card.Content>
            </Card.Root>

            <section class="flex flex-col gap-4 items-stretch w-full">
                <!-- Education Card -->
                <Card.Root class="w-full">
                    <Card.Header
                        class="flex flex-row justify-between items-baseline"
                    >
                        <Card.Title>Education</Card.Title>
                        <Card.Description>
                            <Button variant="ghost" size="icon">+</Button>
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <!-- todo: for each education item, render a special
                             EducationCard that allows edit and delete. 
                             data comes from user/profile endpoint. if no
                             projects, show 'no education added yet' -->
                        <p class="italic opacity-70">No education added yet</p>
                    </Card.Content>
                </Card.Root>

                <!-- Work Experience Card -->
                <Card.Root class="w-full">
                    <Card.Header
                        class="flex flex-row justify-between items-baseline"
                    >
                        <Card.Title>Work Experience</Card.Title>
                        <Card.Description>
                            <Button variant="ghost" size="icon">+</Button>
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <p class="italic opacity-70">
                            No work experience added yet
                        </p>
                    </Card.Content>
                </Card.Root>

                <!-- Projects Card -->
                <Card.Root class="w-full">
                    <Card.Header
                        class="flex flex-row justify-between items-baseline"
                    >
                        <Card.Title>Projects</Card.Title>
                        <Card.Description>
                            <Button variant="ghost" size="icon">+</Button>
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <p class="italic opacity-70">No projects added yet</p>
                    </Card.Content>
                </Card.Root>

                <!-- Skills Card -->
                <Card.Root class="w-full">
                    <Card.Header
                        class="flex flex-row justify-between items-baseline"
                    >
                        <Card.Title>Skills</Card.Title>
                        <Card.Description>
                            <Button variant="ghost" size="icon">+</Button>
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <p class="italic opacity-70">No skills added yet</p>
                    </Card.Content>
                </Card.Root>
            </section>
        </div>
    </div>
</div>