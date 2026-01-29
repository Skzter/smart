<script lang="ts">
    import type { Runner } from "$lib/runner.svelte";
    let { testRunner }: { testRunner: Runner } = $props();
</script>

<div class="flex flex-col h-full overflow-hidden bg-[#0b1220]">
    <!-- LOGS -->
    <div class="flex-1 overflow-auto px-4 py-3 font-mono text-sm">
        {#each testRunner.model.steps as parent}
            <div class="mb-2">
                <div
                    class="flex items-center gap-2
                    {parent.kind === 'group'
                        ? 'text-gray-400 italic'
                        : parent.status === 'failed'
                          ? 'text-red-400'
                          : 'text-gray-200'}"
                >
                    {#if parent.kind === "step"}
                        <span class="w-5 text-center">
                            {#if parent.status === "running"}
                                <span
                                    class="w-4 h-4 border-2 border-blue-400
                                    border-t-transparent rounded-full animate-spin"
                                ></span>
                            {:else if parent.status === "done"}
                                <span class="text-green-400">✓</span>
                            {:else}
                                <span class="text-red-400">✗</span>
                            {/if}
                        </span>
                    {:else}
                        <span class="w-5"></span>
                    {/if}

                    <span>{parent.label}</span>
                </div>

                {#if parent.children?.length}
                    <div class="ml-8 mt-1">
                        {#each parent.children as child}
                            <div
                                class="flex items-center gap-2 mb-1
                                {child.status === 'failed'
                                    ? 'text-red-400'
                                    : 'text-gray-400'}"
                            >
                                <span class="w-5 text-center">
                                    {#if child.status === "running"}
                                        <span
                                            class="w-4 h-4 border-2 border-blue-400
                                            border-t-transparent rounded-full animate-spin"
                                        ></span>
                                    {:else if child.status === "done"}
                                        <span class="text-green-400">✓</span>
                                    {:else}
                                        <span class="text-red-400">✗</span>
                                    {/if}
                                </span>

                                <span>{child.label}</span>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        {/each}
    </div>

    <!-- FOOTER -->
    <div
        class="shrink-0 border-t border-white/10 px-4 py-2 text-sm
        flex items-center justify-between bg-[#0b1220]"
    >
        <div class="flex items-center gap-3">
            {#if testRunner.model.summary.status === "idle"}
                <span class="text-gray-400">inaktiv</span>
            {:else if testRunner.model.summary.status === "running"}
                <span
                    class="w-4 h-4 border-2 border-blue-400
                    border-t-transparent rounded-full animate-spin"
                ></span>
                <span class="text-blue-300">Test läuft…</span>
            {:else if testRunner.model.summary.status === "passed"}
                <span class="text-green-400">✓</span>
                <span class="text-green-300">Test erfolgreich ·</span>
                <span class="text-green-300">
                    {(testRunner.model.summary.durationMs! / 1000).toFixed(1)} s
                </span>
            {:else}
                <span class="text-red-400">✗</span>
                <span class="text-red-300">Test fehlgeschlagen ·</span>
                <span class="text-red-300">
                    {(testRunner.model.summary.durationMs! / 1000).toFixed(1)} s
                </span>
            {/if}
        </div>

        <div class="flex items-center gap-2">
            {#if testRunner.logStatus === "connecting"}
                <span class="w-2 h-2 rounded-full bg-yellow-400 animate-pulse"
                ></span>
                <span class="text-yellow-300">Verbinde Live-Logs…</span>
            {:else if testRunner.logStatus === "connected"}
                <span class="w-2 h-2 rounded-full bg-green-400"></span>
                <span class="text-green-300">Live verbunden</span>
            {:else if testRunner.logStatus === "error"}
                <span class="w-2 h-2 rounded-full bg-red-400"></span>
                <span class="text-red-300">Verbindung fehlgeschlagen</span>
            {:else}
                <span class="w-2 h-2 rounded-full bg-gray-500"></span>
                <span class="text-gray-400">Keine Verbindung</span>
            {/if}
        </div>
    </div>
</div>
