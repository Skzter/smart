<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import type * as Monaco from "monaco-editor";
    import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

    let {
        value = $bindable(),
        language = "typescript",
        options = {},
    }: {
        value?: string;
        language?: string;
        options?: Monaco.editor.IStandaloneEditorConstructionOptions;
    } = $props();

    let container = $state<HTMLElement | undefined>(undefined);
    let editor: Monaco.editor.IStandaloneCodeEditor | undefined = undefined;
    let disposable = $state<Monaco.IDisposable | undefined>(undefined);

    self.MonacoEnvironment = {
        getWorker() {
            return new tsWorker();
        },
    };

    onMount(async () => {
        if (!container) return;

        const monaco = await import("monaco-editor");

        editor = monaco.editor.create(container as HTMLElement, {
            value: value ?? "",
            language,
            automaticLayout: true,
            minimap: { enabled: false },
            theme: "vs-dark",
            // enable word wrap by default; can be overridden via `options` prop
            wordWrap: options.wordWrap ?? "on",
            wrappingStrategy: options.wrappingStrategy ?? "advanced",
            ...options,
        });

        const model = editor.getModel();
        disposable = model?.onDidChangeContent(() => {
            const v = editor?.getValue();
            // update bindable value so parent bindings update
            value = v;
        });
    });

    onDestroy(() => {
        disposable?.dispose();
        editor?.dispose();
    });

    $effect(() => {
        if (!editor) return;
        if (value !== undefined && editor.getValue() !== value) {
            const pos = editor.getPosition();
            editor.setValue(value);
            if (pos) editor.setPosition(pos);
        }
    });
</script>

<div bind:this={container} class="h-full min-h-[200px] w-full"></div>

<style>
    :global(.monaco-editor) {
        height: 100%;
    }
</style>
