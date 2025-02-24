<script lang="ts">
    import { enhance } from "$app/forms";
    import { goto } from '$app/navigation'

    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";

    export let objectID: string; // temporarily not used; will be used for db operations later
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp 
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;

    async function deleteApplication() {
        const formData = new FormData();
        formData.append("id", objectID);
        formData.append("company", company);
        formData.append("role", role);
        formData.append("appliedDate", String(appliedDate));
        formData.append("location", location);
        formData.append("status", status);
        formData.append("link", link || "");

        const response = await fetch(`/dashboard?/delete`, {
            method: "POST",
            body: formData,
        });

        if (!response.ok) {
            console.error("Failed to delete application");
        } else {
            console.log("Application deleted successfully");
            goto("/dashboard");
        }
    }
</script>

<AlertDialog.Root>
    <AlertDialog.Trigger asChild let:builder>
        <Button
            builders={[builder]}
            variant="outline"
            class="text-primary focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 px-4 py-2 text-xs flex-grow sm:border-0 sm:shadow-none sm:hover:bg-transparent sm:hover:bg-accent"
        >
            Edit
        </Button>
    </AlertDialog.Trigger>
    <AlertDialog.Content>
        <AlertDialog.Header class="text-left">
            <AlertDialog.Title>Edit Application</AlertDialog.Title>
            <AlertDialog.Description>
                Update your application. Unmodified fields will
                remain unchanged.
            </AlertDialog.Description>
            <form
                action="/dashboard?/editapplication"
                method="POST"
                class="flex flex-col gap-2 w-full"
                use:enhance={() => {
                    return ({ update }) => {
                        update();
                    };
                }}
            >
                <!-- if any field is left empty, value will be set to the current value else overridden by the new value -->
                <!-- hidden id field and old statuses. old status are sent for rollback purposes -->
                <input type="hidden" name="id" value={objectID} />
                <input type="hidden" name="oldCompany" value={company} />
                <input type="hidden" name="oldRole" value={role} />
                <input type="hidden" name="oldLocation" value={location} />
                <input type="hidden" name="oldAppliedDate" value={appliedDate} />
                <input type="hidden" name="oldLink" value={link} />
                <div
                    class="grid grid-cols-[2fr_5fr] sm:grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                >
                    <Label
                        for="company"
                        class="text-sm text-gray-500 font-light"
                        >Company</Label
                    >
                    <Input
                        type="text"
                        name="company"
                        placeholder="Company"
                        value={company}
                    />
                </div>
                <div
                    class="grid grid-cols-[2fr_5fr] sm:grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                >
                    <Label
                        for="role"
                        class="text-sm text-gray-500 font-light"
                        >Role</Label
                    >
                    <Input
                        type="text"
                        name="role"
                        placeholder="Role"
                        value={role}
                    />
            </div>
                <div
                    class="grid grid-cols-[2fr_5fr] sm:grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                >
                    <Label
                        for="location"
                        class="text-sm text-gray-500 font-light"
                        >Location</Label
                    >
                    <Input
                        type="text"
                        name="location"
                        placeholder="Location"
                        value={location}
                    />
                </div>
                <div
                    class="grid grid-cols-[2fr_5fr] sm:grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                >
                    <Label
                        for="link"
                        class="text-sm text-gray-500 font-light"
                        >Link</Label
                    >
                    <Input
                        type="text"
                        name="link"
                        placeholder="Link"
                        value={link}
                    />
                </div>
                <div
                    class="grid grid-cols-[2fr_5fr] sm:grid-cols-[1fr_5fr] w-full items-center gap-1.5"
                >
                    <Label
                        for="appliedDate"
                        class="text-sm text-gray-500 font-light"
                        >Applied Date</Label
                    >
                    <Input
                        type="date"
                        name="appliedDate"
                        placeholder="Applied Date"
                        value={appliedDate}
                    />
                </div>
                <AlertDialog.Footer>
                    <AlertDialog.Action type="submit" class="w-full sm:w-auto h-9 px-4 py-2 mt-2 sm:mt-0">
                        Save
                    </AlertDialog.Action>

                    <div class="flex gap-2">
                        
                        <!-- <AlertDialog.Trigger asChild let:builder>
                            <Button builders={[builder]} variant="outline" class="w-full sm:hidden text-red-500 focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 px-4 py-2 mt-2 sm:mt-0">
                                Delete
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
                                on:click={deleteApplication}
                              >
                                Continue
                                </AlertDialog.Action>
                            </AlertDialog.Footer>
                          </AlertDialog.Content> -->

                        <AlertDialog.Cancel class="w-full">Cancel</AlertDialog.Cancel>

                        <AlertDialog.Root>
                            <AlertDialog.Trigger asChild let:builder>
                                <Button builders={[builder]} variant="outline" class="w-full sm:hidden text-red-500 focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:pointer-events-none disabled:opacity-50 border-input bg-background hover:bg-accent hover:text-accent-foreground border shadow-sm h-9 px-4 py-2 mt-2 sm:mt-0">
                                    Delete
                                </Button>
                            </AlertDialog.Trigger>
                            <AlertDialog.Content>
                                    <AlertDialog.Header>
                                        <AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
                                        <AlertDialog.Description>
                                            This action cannot be undone. This will permanently delete this application
                                            data from our servers.
                                        </AlertDialog.Description>
                                    </AlertDialog.Header>
                                    <AlertDialog.Footer>
                                        <AlertDialog.Action
                                            on:click={deleteApplication}
                                        >
                                            Continue
                                        </AlertDialog.Action>
                                        <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
                                    </AlertDialog.Footer>
                              </AlertDialog.Content>
                        </AlertDialog.Root>
                    </div>
                </AlertDialog.Footer>
            </form>
        </AlertDialog.Header>
    </AlertDialog.Content>
</AlertDialog.Root>