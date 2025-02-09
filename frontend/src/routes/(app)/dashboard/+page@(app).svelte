<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Separator } from "$lib/components/ui/separator";
  import { Job } from "$lib/components/Job";
  import * as AlertDialog from "$lib/components/ui/alert-dialog";

  import { enhance } from "$app/forms";
  import type { PageData } from "./$types";

  let open = false;

  export let data: PageData;
</script>

<div
  class="flex flex-col justify-start gap-4 items-stretch w-full sm:w-5/6 mx-auto h-full"
>
  <div class="p-3 my-10">
    <div
      class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-center w-full sm:min-w-[72vw]"
    >
      <div class="flex flex-row flex-grow gap-4 items-center w-full sm:w-auto">
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
              <AlertDialog.Title>Add Application</AlertDialog.Title>
              <AlertDialog.Description>
                Add a new application to your list. Defaults to 'Applied'
                status.
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
                <Input type="text" name="role" placeholder="Role" required />
                <Input
                  type="text"
                  name="company"
                  placeholder="Company"
                  required
                />
                <Input
                  type="text"
                  name="location"
                  placeholder="Location"
                  required
                />
                <Input type="text" name="link" placeholder="Link (Optional)" />
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

        <Separator orientation="vertical" class="h-6" />
        <Input type="text" placeholder="Search for roles or companies" />
      </div>
      <div class="flex flex-row gap-2 sm:gap-4 items-center w-full sm:w-auto">
        <div class="flex items-center gap-0 sm:gap-2 -ml-3 sm:ml-0">
          <Button variant="ghost" type="button" class="text-xs sm:text-sm"
            >Applied from</Button
          >
          <Separator orientation="horizontal" class="w-2 sm:w-3" />
          <Button variant="ghost" type="button" class="text-xs sm:text-sm"
            >Applied until</Button
          >
        </div>
        <Separator orientation="vertical" class="h-4 sm:h-6" />
        <Button variant="ghost" type="button" class="text-xs sm:text-sm"
          >Job Type</Button
        >
        <Separator orientation="vertical" class="h-4 sm:h-6" />
        <Button variant="ghost" type="button" class="text-xs sm:text-sm"
          >Status</Button
        >
      </div>
    </div>

    <div>
      {#each data.applications as job (job.id)}
        <Job
          id={job.id}
          company={job.company}
          role={job.role}
          appliedDate={job.appliedDate}
          location={job.location}
          status={job.status}
          link={job.link}
        />
      {/each}
    </div>
  </div>
</div>
