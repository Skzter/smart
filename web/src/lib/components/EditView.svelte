<script lang="ts">
    import Code from "./Code.svelte";
    import RunButton from "./RunButton.svelte";
    import SaveButton from "./SaveButton.svelte";

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
    <div class="overflow-y-auto">
        <Code {code} />
    </div>
    <div
        class="bg-gray-100 flex items-center justify-center border-l overflow-y-auto p-6"
    >
        <div class="w-full space-y-4">
            <div class="bg-gray-50 rounded-lg p-6 border border-gray-200">
                <h3 class="text-sm font-semibold mb-4 text-gray-600">
                    Test Information
                </h3>
                <div class="space-y-3">
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600 text-sm">Zeilen:</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600 text-sm">Zeichen:</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600 text-sm">Status:</span>
                    </div>
                </div>
            </div>
            <div class="bg-gray-50 rounded-lg p-6 border border-gray-200">
                <h3 class="text-sm font-semibold mb-4 text-gray-600">
                    Schnellaktionen
                </h3>
                <div class="space-y-3">
                    <RunButton bind:isLoading bind:activeTab bind:testResult />
                    <SaveButton {code} />
                </div>
            </div>
        </div>
    </div>
</div>
