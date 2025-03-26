<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Badge } from "$lib/components/ui/badge";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import * as Card from "$lib/components/ui/card/index.js";

    import { Moon, X } from "lucide-svelte";
    import { SunMedium } from "lucide-svelte";

    import Chart from "chart.js/auto";

    import { ModeWatcher } from "mode-watcher";
    import { toggleMode } from "mode-watcher";

    import type { PageData } from "./$types";
    import { onMount } from "svelte";

    import { formatDateWithSeconds } from "$lib/utils/date";

    function signOut() {
        window.location.href = "/auth/google/logout";
    }

    // method not allowed error with method DELETE
    // even though cors is set up correctly so just post with empty body
    async function deleteAccount() {
        const formData = new FormData();
        const response = await fetch("?/delete", {
            method: "POST",
            body: formData,
            headers: {
		        'x-sveltekit-action': 'true'
	        }
        });

        if (!response.ok) {
            console.error("Failed to delete account");
        }

        signOut();
    }

    let chartCanvas: HTMLCanvasElement;
    let chart: Chart;

    // create a chart on mount
    onMount(() => {
        if (data.analytics?.monthly_trends && chartCanvas) {
            //@ts-ignore
            const months = data.analytics.monthly_trends.map((item) => item.Month);
            //@ts-ignore
            const applications = data.analytics.monthly_trends.map((item) => item.Applications);
            //@ts-ignore
            const interviews = data.analytics.monthly_trends.map((item) => item.Interviews);
            //@ts-ignore
            const offers = data.analytics.monthly_trends.map((item) => item.Offers);

            chart = new Chart(chartCanvas, {
                type: "line",
                data: {
                    labels: months,
                    datasets: [
                        {
                            label: "Applications",
                            data: applications,
                            backgroundColor: "rgba(59, 130, 246, 0.1)",
                            borderColor: "rgb(59, 130, 246)",
                            borderWidth: 2,
                            pointRadius: 4,
                            pointHoverRadius: 6,
                            tension: 0.3,
                            fill: true,
                        },
                        {
                            label: "Interviews",
                            data: interviews,
                            backgroundColor: "rgba(168, 85, 247, 0.1)",
                            borderColor: "rgb(168, 85, 247)",
                            borderWidth: 2,
                            pointRadius: 4,
                            pointHoverRadius: 6,
                            tension: 0.3,
                            fill: true,
                        },
                        {
                            label: "Offers",
                            data: offers,
                            backgroundColor: "rgba(34, 197, 94, 0.1)",
                            borderColor: "rgb(34, 197, 94)",
                            borderWidth: 2,
                            pointRadius: 4,
                            pointHoverRadius: 6,
                            tension: 0.3,
                            fill: true,
                        },
                    ],
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                precision: 0,
                            },
                        },
                        x: {
                            grid: {
                                display: false,
                            },
                        },
                    },
                    elements: {
                        line: {
                            tension: 0.3, // Smooth curves
                        },
                    },
                    plugins: {
                        legend: {
                            position: "top",
                            labels: {
                                usePointStyle: true,
                                boxWidth: 16,
                            },
                        },
                        tooltip: {
                            backgroundColor: "rgba(0, 0, 0, 0.7)",
                            padding: 10,
                            titleFont: {
                                size: 14,
                            },
                            bodyFont: {
                                size: 13,
                            },
                        },
                    },
                },
            });
        }

        // clean up chart on unmount
        return () => {
            if (chart) {
                chart.destroy();
            }
        };
    });

    export let data: PageData;
</script>

<ModeWatcher />
<div
    class="px-6 sm:px-8 flex flex-col justify-start gap-4 items-stretch w-full mx-auto h-full my-12 dark:brightness-[0.9]"
>
    <div
        class="flex flex-col sm:flex-row justify-between gap-2 sm:gap-4 items-start w-full"
    >
        <div class="w-full sm:w-fit flex flex-col items-baseline justify-between mb-2">
            <h2 class="text-2xl font-bold tracking-tight mb-4">Manage</h2>
            <Card.Root class="w-full">
                <Card.Header>
                    <Card.Title>{data.email}</Card.Title>
                    <Card.Description
                        >Total Applications: {data.applicationsCount}</Card.Description
                    >
                </Card.Header>
                <Card.Content>
                    <div class="grid grid-cols-2 gap-2">
                        <div class="col-span-1 flex justify-center">
                            <Button
                                variant="outline"
                                class="w-full"
                                on:click={signOut}
                            >
                                Sign out
                            </Button>
                        </div>
                        <div class="col-span-1 flex justify-center">
                            <Button
                                variant="outline"
                                class="w-full"
                                on:click={toggleMode}
                            >
                                <SunMedium class="dark:hidden mr-2 h-6 w-6" />
                                <Moon class="hidden dark:block mr-2 h-6 w-6" />
                                Toggle theme
                            </Button>
                        </div>
                    </div>
                    <div class="grid grid-cols-1 mt-2">
                        <AlertDialog.Root>
                            <AlertDialog.Trigger asChild let:builder>
                                <Button
                                    builders={[builder]}
                                    variant="outline"
                                    class="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-300 dark:hover:text-red-700"
                                >
                                    Delete account
                                </Button>
                            </AlertDialog.Trigger>
                            <AlertDialog.Content>
                                <AlertDialog.Header>
                                    <AlertDialog.Title
                                        >Are you absolutely sure?</AlertDialog.Title
                                    >
                                    <AlertDialog.Description>
                                        This action cannot be undone. This will
                                        permanently delete your account and
                                        remove your data from our servers.
                                    </AlertDialog.Description>
                                </AlertDialog.Header>
                                <AlertDialog.Footer>
                                    <AlertDialog.Cancel
                                        >Cancel</AlertDialog.Cancel
                                    >
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
        </div>

        <section class="flex flex-col gap-4 items-stretch w-full sm:w-3/4 mb-4">
            <div class="flex flex-col sm:flex-row items-left gap-2 sm:gap-0 sm:items-center justify-between">
                <h2 class="text-2xl font-bold tracking-tight">
                    Application Analytics
                </h2>
                <Badge class="w-fit" variant="outline">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        class="mr-2 h-4 w-4"
                        ><path d="M3 12a9 9 0 1 0 18 0 9 9 0 0 0-18 0"
                        ></path><path d="M3 12h18"></path></svg
                    >
                    Last updated: {formatDateWithSeconds(data.analytics?.last_updated)}
                </Badge>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                <!-- Application Velocity -->
                <Card.Root>
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium"
                            >Application Velocity</Card.Description
                        >
                    </Card.Header>
                    <Card.Content>
                        <div class="flex items-center justify-between">
                            <div class="text-2xl font-bold">
                                {data.analytics?.application_velocity_trend > 0
                                    ? "+"
                                    : ""}{data.analytics
                                    ?.application_velocity_trend || 0}
                            </div>
                            <div
                                class={`flex items-center ${
                                    data.analytics?.application_velocity_trend > 0
                                    ? "text-green-500"
                                    : data.analytics?.application_velocity_trend < 0
                                        ? "text-red-500"
                                        : "text-muted-foreground"}`}
                            >
                                {#if data.analytics?.application_velocity_trend > 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                        ><path d="m5 12 7-7 7 7"></path><path
                                            d="M5 19h14"
                                        ></path></svg
                                    >
                                {:else if data.analytics?.application_velocity_trend < 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                        ><path d="m5 12 7 7 7-7"></path><path
                                            d="M5 5h14"
                                        ></path></svg
                                    >
                                {:else}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                    >
                                        <path d="M5 9h14"></path>
                                        <path d="M5 15h14"></path>
                                  </svg>
                                {/if}
                                <span
                                    >{Math.abs(
                                        data.analytics?.application_velocity || 0)} apps sent</span
                                >
                            </div>
                        </div>
                        <p class="text-sm text-muted-foreground mt-2">
                            {#if data.analytics?.application_velocity_trend > 0}
                                More
                            {:else if data.analytics?.application_velocity_trend < 0}
                                Fewer
                            {:else}
                                Same number of
                            {/if} applications than previous 30 day period
                        </p>
                    </Card.Content>
                </Card.Root>

                <!-- Resume Effectiveness -->
                <Card.Root>
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium"
                            >Resume Effectiveness</Card.Description
                        >
                    </Card.Header>
                    <Card.Content>
                        <div class="flex items-center justify-between">
                            <div class="text-2xl font-bold">
                                {data.analytics?.resume_effectiveness_trend > 0
                                    ? "+"
                                    : ""}{data.analytics
                                    ?.resume_effectiveness_trend || 0}
                            </div>
                            <div
                                class={`flex items-center ${
                                    data.analytics?.resume_effectiveness_trend > 0
                                    ? "text-green-500"
                                    : data.analytics?.resume_effectiveness_trend < 0
                                        ? "text-red-500"
                                        : "text-muted-foreground"}`}
                            >
                                {#if data.analytics?.resume_effectiveness_trend > 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                    >
                                        <path d="m5 12 7-7 7 7"></path>
                                        <path d="M5 19h14"></path>
                                    </svg>
                                {:else if data.analytics?.resume_effectiveness_trend < 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                        ><path d="m5 12 7 7 7-7"></path><path
                                            d="M5 5h14"
                                        ></path></svg
                                    >
                                {:else}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                 >
                                    <path d="M5 9h14"></path>
                                    <path d="M5 15h14"></path>
                                </svg>
                                {/if}
                                <span>{Math.abs(data.analytics?.resume_effectiveness || 0)} interviews</span>
                            </div>
                        </div>
                        <p class="text-sm text-muted-foreground mt-2">
                            {#if data.analytics?.resume_effectiveness_trend > 0}
                                More
                            {:else if data.analytics?.resume_effectiveness_trend < 0}
                                Fewer
                            {:else}
                                Same number of
                            {/if} interviews than previous 30 day period
                        </p>
                    </Card.Content>
                </Card.Root>

                <!-- Interview Success Rate -->
                <Card.Root>
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium">
                            Offer Conversion
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <div class="flex items-center justify-between">
                            <div class="text-2xl font-bold">
                                {data.analytics?.interview_effectiveness_trend > 0
                                    ? "+"
                                    : ""}
                                {data.analytics?.interview_effectiveness_trend || 0}
                            </div>
                            <div
                                class={`flex items-center ${
                                    data.analytics?.interview_effectiveness_trend > 0 
                                    ? "text-green-500" 
                                    : data.analytics?.interview_effectiveness_trend < 0 
                                        ? "text-red-500" 
                                        : "text-muted-foreground"
                                }`}
                            >
                                {#if data.analytics?.interview_effectiveness_trend > 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                    >
                                        <path d="m5 12 7-7 7 7"></path>
                                        <path d="M5 19h14"></path>
                                    </svg>
                                {:else if data.analytics?.interview_effectiveness_trend < 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                    >
                                        <path d="m5 12 7 7 7-7"></path>
                                        <path d="M5 5h14"></path>
                                    </svg>
                                {:else}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4 mr-1"
                                    >
                                        <path d="M5 9h14"></path>
                                        <path d="M5 15h14"></path>
                                    </svg>
                                {/if}
                                <span>{data.analytics?.interview_effectiveness || 0} offers</span>
                            </div>
                        </div>
                        <p class="text-sm text-muted-foreground mt-2">
                            {#if data.analytics?.interview_effectiveness_trend > 0}
                                More
                            {:else if data.analytics?.interview_effectiveness_trend < 0}
                                Fewer
                            {:else}
                                Same number of
                            {/if} offers than previous 30 day period
                        </p>
                    </Card.Content>
                </Card.Root>

                <!-- Average Response Time -->
                <Card.Root>
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium">
                            Average Response Time
                        </Card.Description>
                    </Card.Header>
                    <Card.Content>
                        <div class="flex items-center justify-between">
                            <div class="text-2xl font-bold">
                                <!-- if nil then X.X -->
                                {data.analytics?.avg_response_time != null
                                    ? data.analytics.avg_response_time.toFixed(1)
                                    : "X.X"
                                } 
                                days
                            </div>
                            <div
                                class={`flex items-center ${data.analytics?.avg_response_time_trend < 0
                                    ? "text-green-500"
                                    : data.analytics?.avg_response_time_trend > 0
                                        ? "text-red-500"
                                        : "text-muted-foreground"}`
                                }
                            >
                                {#if data.analytics?.avg_response_time_trend < 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4"
                                    >
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <polyline points="12 6 12 12 16 14"></polyline>
                                    </svg>
                                {:else if data.analytics?.avg_response_time_trend > 0}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4"
                                        ><circle cx="12" cy="12" r="10"
                                        ></circle><polyline
                                            points="12 6 12 12 8 14"
                                        ></polyline></svg
                                    >
                                {:else}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="h-4 w-4"
                                        ><circle cx="12" cy="12" r="10"
                                        ></circle><polyline
                                            points="12 6 12 12 16 14"
                                        ></polyline></svg
                                    >
                                {/if}
                            </div>
                        </div>
                        <p class="text-sm text-muted-foreground mt-2">
                            {#if data.analytics?.avg_response_time_trend < 0}
                                Faster
                            {:else if data.analytics?.avg_response_time_trend > 0}
                                Slower
                            {:else}
                                Same response time
                            {/if} than previous 30 day period ({data.analytics?.avg_response_time_trend != null 
                                    ? Math.abs(data.analytics.avg_response_time_trend).toFixed(1)
                                    : 'X.X'}
                                days)
                        </p>
                    </Card.Content>
                </Card.Root>

                <!-- Application Status Distribution -->
                <Card.Root class="sm:col-span-2">
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium"
                            >Application Status Distribution</Card.Description
                        >
                    </Card.Header>
                    <Card.Content>
                        <div class="space-y-2">
                            <!-- Calculate total for percentages -->
                            <!-- unfortunately sometimes it doesnt add up to a perfect 100 due to rounding errors lol -->
                            {#if data.applicationsCount > 0}
                                {@const total = data.applicationsCount}
                                {@const appliedPercent = Math.round(
                                    (data.analytics.applied_count / total) *
                                        100,
                                )}
                                {@const screenPercent = Math.round(
                                    (data.analytics.screen_count / total) * 100,
                                )}
                                {@const interviewingPercent = Math.round(
                                    (data.analytics.interviewing_count /
                                        total) *
                                        100,
                                )}
                                {@const offerPercent = Math.round(
                                    (data.analytics.offer_count / total) * 100,
                                )}
                                {@const rejectedPercent = Math.round(
                                    (data.analytics.rejected_count / total) *
                                        100,
                                )}
                                {@const ghostedPercent = Math.round(
                                    (data.analytics.ghosted_count / total) *
                                        100,
                                )}

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-blue-500"
                                        ></div>
                                        <span class="text-sm">Applied</span>
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{appliedPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-blue-500 h-full"
                                                style="width: {appliedPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-purple-500"
                                        ></div>
                                        <span class="text-sm">Screen</span>
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{screenPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-purple-500 h-full"
                                                style="width: {screenPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-amber-500"
                                        ></div>
                                        <span class="text-sm">Interviewing</span
                                        >
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{interviewingPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-amber-500 h-full"
                                                style="width: {interviewingPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-green-500"
                                        ></div>
                                        <span class="text-sm">Offer</span>
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{offerPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-green-500 h-full"
                                                style="width: {offerPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-red-500"
                                        ></div>
                                        <span class="text-sm">Rejected</span>
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{rejectedPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-red-500 h-full"
                                                style="width: {rejectedPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="h-3 w-3 rounded-full bg-gray-500"
                                        ></div>
                                        <span class="text-sm">Ghosted</span>
                                    </div>
                                    <div class="flex items-center gap-1">
                                        <span>{ghostedPercent || 0}%</span>
                                        <div
                                            class="w-24 h-2 bg-muted rounded-full overflow-hidden"
                                        >
                                            <div
                                                class="bg-gray-500 h-full"
                                                style="width: {ghostedPercent ||
                                                    0}%"
                                            ></div>
                                        </div>
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </Card.Content>
                </Card.Root>
                <!-- monthly trends -->
                <Card.Root class="col-span-1 sm:col-span-2 lg:col-span-3">
                    <Card.Header class="pb-2">
                        <Card.Description class="text-sm font-medium"
                            >Monthly Trends (1 year ago - now)</Card.Description
                        >
                    </Card.Header>
                    <Card.Content class="h-[300px]">
                        <div class="w-full h-full">
                            <canvas bind:this={chartCanvas}></canvas>
                        </div>
                    </Card.Content>
                </Card.Root>
            </div>
        </section>
    </div>
</div>
