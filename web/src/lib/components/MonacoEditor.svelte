<script lang="ts">
    import { editor, type IDisposable } from "monaco-editor";
    import { onMount, onDestroy } from "svelte";
    import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

    let {
        value = $bindable(),
        options,
        class: className = "",
    }: {
        value: string;
        options?: {
            maxHeight?: number;
            useTextHeight?: boolean;
        };
        class?: string;
    } = $props();

    let container = document.createElement("div");
    let codeEditor = $state<editor.IStandaloneCodeEditor | undefined>(
        undefined,
    );
    let disposable = $state<IDisposable | undefined>(undefined);

    self.MonacoEnvironment = {
        getWorker() {
            return new tsWorker();
        },
    };

    onMount(async () => {
        if (!container) return;

        const monaco = await import("monaco-editor");

        codeEditor = monaco.editor.create(container as HTMLElement, {
            value: value ?? "",
            language: "typescript",
            minimap: { enabled: false },
            theme: "vs-dark",
            wordWrap: "on",
            allowOverflow: true,
            lineHeight: 1.5,
            scrollBeyondLastLine: false,
        });

        const model = codeEditor.getModel();
        disposable = model?.onDidChangeContent(() => {
            const v = codeEditor?.getValue();
            // update bindable value so parent bindings update
            value = v ?? "";
        });
    });

    onDestroy(() => {
        disposable?.dispose();
        codeEditor?.dispose();
    });

    $effect(() => {
        if (!codeEditor) return;
        if (value !== undefined && codeEditor.getValue() !== value) {
            const pos = codeEditor.getPosition();
            codeEditor.setValue(value);
            if (pos) codeEditor.setPosition(pos);
        }

        let height: number;
        if (options?.useTextHeight) {
            height = codeEditor.getContentHeight();
        } else {
            height = container.clientHeight;
        }
        console.log(Math.min(height, options?.maxHeight ?? 10000));
        codeEditor.layout({
            height: Math.min(height, options?.maxHeight ?? 10000),
            width: container.clientWidth,
        });
    });
</script>

<div bind:this={container} class={className}></div>

<style>
    :global(.monaco-editor) {
        height: 100%;
    }
</style>
