<script lang="ts">
    import { enhance } from "$app/forms";
    import { goto } from '$app/navigation'

    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";

    export let objectID: string;
    export let company: string;
    export let role: string;
    export let appliedDate: number; // raw unix timestamp 
    export let location: string;
    export let status: string;
    export let link: string | undefined | null;

    export let onUpdateSuccess: (updatedJob: any) => void = () => {};

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

    async function handleEditSubmit(e: Event) {
        e.preventDefault(); // Prevent form from submitting normally
        const form = e.target as HTMLFormElement;
        const formData = new FormData(form);
        
        // Add required fields
        formData.append("id", objectID);
        
        console.log("Submitting edit form...");
        
        const response = await fetch("/dashboard?/editapplication", {
            method: "POST",
            body: formData
        });
        
        if (response.ok) {
            console.log("Edit successful");
            
            // Get form data and create updated job object
            const updatedJob = {
                company: (formData.get('company') as string) || company,
                role: (formData.get('role') as string) || role,
                location: (formData.get('location') as string) || location,
                link: (formData.get('link') as string) || link,
                appliedDate: formData.get('appliedDate')
                    ? convertLocalDateToTimestamp(formData.get('appliedDate') as string)
                    : appliedDate,
                status: (formData.get('status') as string) || status,
            };
            
            // Call the callback with updated data
            console.log("Calling onUpdateSuccess with:", updatedJob);
            onUpdateSuccess(updatedJob);
            
            // Close the dialog
            const dialogElement = document.querySelector('[data-state="open"]');
            if (dialogElement) {
                const cancelButton = dialogElement.querySelector('[data-dialog-close]');
                if (cancelButton instanceof HTMLElement) {
                    cancelButton.click();
                }
            }
        } else {
            console.error("Failed to update application");
        }
    }

    function convertLocalDateToTimestamp(dateString: string): number {
    // Parse the input date string (YYYY-MM-DD)
    const [year, month, day] = dateString.split('-').map(Number);
    
    // Create a date using local timezone (months are 0-indexed in JS Date)
    const date = new Date(year, month - 1, day, 12, 0, 0);
    
    // Get the timestamp in seconds (not milliseconds)
    return Math.floor(date.getTime() / 1000);
}

    // format to mm-dd-yyyy
    function formatDate(timestamp: number): string {
        if (!timestamp) return "";

        const date = new Date(timestamp * 1000);
        if (isNaN(date.getTime())) return "";

        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");
        const year = date.getFullYear();

        return `${year}-${month}-${day}`;
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
                on:submit={handleEditSubmit}
                class="flex flex-col gap-2 w-full"
            >
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
                        value={formatDate(appliedDate)}
                    />
                </div>

                <input type="hidden" name="id" value={objectID} />

                <AlertDialog.Footer>
                    <AlertDialog.Action type="submit" class="w-full sm:w-auto h-9 px-4 py-2 mt-2 sm:mt-0">
                        Save
                    </AlertDialog.Action>

                    <div class="flex gap-2">
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