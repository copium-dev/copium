
<script>
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import { Input } from "$lib/components/ui/input";
    import { Button } from "$lib/components/ui/button";
    
    import { enhance } from "$app/forms";

    let open = false;
</script>

<AlertDialog.Root bind:open>
    <AlertDialog.Trigger asChild let:builder>
        <Button
            builders={[builder]}
            variant="outline"
            class="w-16"
            on:click={() => {
                open = true;
            }}
        >
            Add
        </Button>
    </AlertDialog.Trigger>
    <AlertDialog.Content>
        <AlertDialog.Header>
            <AlertDialog.Title
                >Add Application</AlertDialog.Title
            >
            <AlertDialog.Description>
                Add a new application to your list. Defaults to
                'Applied' status.
            </AlertDialog.Description>
            <form
                action="?/add"
                method="POST"
                class="flex flex-col gap-2 w-full"
                use:enhance={(data) => {
                    return async ({ update }) => {
                        if (data.formElement.checkValidity()) {
                            open = false;
                            update();
                        } else {
                            // Report validity errors so the dialog remains open.
                            data.formElement.reportValidity();
                        }
                    };
                }}
            >
                <Input
                    type="text"
                    name="company"
                    placeholder="Company"
                    required
                />
                <Input
                    type="text"
                    name="role"
                    placeholder="Role"
                    required
                />
                <Input
                    type="text"
                    name="location"
                    placeholder="Location"
                    required
                />
                <Input
                    type="text"
                    name="link"
                    placeholder="Link (Optional)"
                />
                <Input
                    type="date"
                    name="appliedDate"
                    placeholder="Applied Date"
                    required
                />
                <AlertDialog.Footer>
                    <AlertDialog.Cancel
                        on:click={() => {
                            open = false;
                        }}
                    >
                        Cancel
                    </AlertDialog.Cancel>
                    <AlertDialog.Action asChild>
                        <Button type="submit">Add</Button>
                    </AlertDialog.Action>
                </AlertDialog.Footer>
            </form>
        </AlertDialog.Header>
    </AlertDialog.Content>
</AlertDialog.Root>
