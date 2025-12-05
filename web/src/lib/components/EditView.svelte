<script lang="ts">
    import MonacoEditor from "./MonacoEditor.svelte";
    import { Play, Save } from "@lucide/svelte";
    import { runContainer } from "$lib/api";
    import { chat, user } from "$lib/shared.svelte";

    let {
        testResult = $bindable(),
        code = $bindable(),
        isLoading = $bindable(),
        activeTab = $bindable(),
    }: {
        testResult: string;
        code: string;
        isLoading: boolean;
        activeTab: string;
    } = $props();
</script>

<div class="flex-1 grid gap-0 h-full" style="grid-template-columns: 70% 30%">
    <div class="overflow-y-auto h-full">
        <MonacoEditor bind:value={code} language="typescript" />
    </div>
    <div
        class="bg-gray-300 flex items-center justify-center border-l overflow-y-auto p-6"
    >
        <div class="w-full space-y-4">
            <div class="bg-gray-50 rounded-lg p-6 border border-gray-200">
                <h1 class="text-md font-semibold mb-4 text-gray-600">
                    Test Information
                </h1>
                <div class="space-y-3">
                    <p class="text-gray-600 text-sm">Zeilen:</p>
                    <p class="text-gray-600 text-sm">Zeichen:</p>
                    <p class="text-gray-600 text-sm">Status:</p>
                </div>
            </div>
            <div class="bg-gray-50 rounded-lg p-6 border border-gray-200">
                <h1 class="text-md font-semibold mb-4 text-gray-600">
                    Schnellaktionen
                </h1>
                <div class="flex flex-col text-black">
                    <RunButton
                        classes="flex justify-start"
                        variant="ghost"
                        size="sm"
                        bind:isLoading
                        bind:activeTab
                        bind:testResult
                    />
                    <SaveButton
                        classes="flex justify-start"
                        variant="ghost"
                        size="sm"
                        {code}
                    />
                </div>
            </div>
        </div>
    </div>
</div>
