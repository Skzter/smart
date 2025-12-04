<script lang="ts">
    import Code from "./Code.svelte";
    import { Play, Save } from "@lucide/svelte";
    import { runContainer } from "$lib/api";


    let {
        code,
        userId,
        testId,
        sessionId,
        onRunClick,
    }: {
        code: string;
        userId: string;
        testId: string;
        sessionId: string;
        onRunClick: (result: string) => void;
    } = $props();

    let isLoading = $state(false);

    async function handleTestRun() {
        isLoading = true;
        try {
            const response = await runContainer({
                userId,
                testId,
                sessionId,
            });
            onRunClick(response.result);
        } catch (error) {
            console.error('Error running test:', error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="flex-1 grid gap-0 h-full" style="grid-template-columns: 70% 30%">
  <div class="overflow-y-auto">
    <Code {code} />
  </div>
  <div class="bg-gray-100 flex items-center justify-center border-l overflow-y-auto p-6">
    <div class="w-full space-y-4">
      <div class="bg-gray-50 rounded-lg p-6 border border-gray-200">
        <h3 class="text-sm font-semibold mb-4 text-gray-600">Test Information</h3>
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
        <h3 class="text-sm font-semibold mb-4 text-gray-600">Schnellaktionen</h3>
        <div class="space-y-3">
          <button onclick={handleTestRun} disabled={isLoading} class="w-full flex items-center gap-3 p-3 hover:bg-gray-100 rounded cursor-pointer transition disabled:opacity-50 disabled:cursor-not-allowed">
            <Play class="w-4 h-4 text-gray-800" />
            <span class="text-gray-800">{isLoading ? 'Lädt...' : 'Test ausführen'}</span>
          </button>
          <button class="w-full flex items-center gap-3 p-3 hover:bg-gray-100 rounded cursor-pointer transition opacity-50">
            <Save class="w-4 h-4 text-gray-800" />
            <span class="text-gray-800">Speichern</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
