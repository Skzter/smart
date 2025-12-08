// MonacoModel.ts
import * as monaco from "monaco-editor";
import { SvelteMap } from "svelte/reactivity";
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

/**
 * Represents a single Monaco editor model instance and its reactive Svelte 5 state.
 */
export class MonacoModel {
    // Svelte 5 $state creates a reactive proxy for the content property.
    // This allows Svelte components to read/write to it reactively.
    public language: string = "typescript";
    public id: string;

    // Non-reactive reference to the actual Monaco model object
    private monacoModel: monaco.editor.ITextModel;
    private contentChangeListener: monaco.IDisposable | null = null;

    constructor(
        id: string,
        initialContent: string,
        contentChangeListener: (
            e: monaco.editor.IModelContentChangedEvent,
        ) => void,
    ) {
        this.id = id;

        // 1. Create the Monaco model
        const uri = monaco.Uri.file(id);
        this.monacoModel = monaco.editor.createModel(
            initialContent,
            this.language,
            uri,
        );

        this.contentChangeListener = this.monacoModel.onDidChangeContent(
            contentChangeListener,
        );
    }

    /**
     * Retrieves the internal Monaco ITextModel object.
     */
    public getMonacoModel(): monaco.editor.ITextModel {
        return this.monacoModel;
    }

    /**
     * Cleans up the Monaco model and its subscriptions.
     */
    public dispose(): void {
        this.contentChangeListener?.dispose();
        this.monacoModel.dispose();
    }
}

// Map to hold references to all managed models
type ModelMap = SvelteMap<string, MonacoModel>;

/**
 * Manages a static, lazily-initialized Monaco Editor instance and its models.
 */
export class ModelManager {
    private static models: ModelMap = new SvelteMap();
    private static isInitializing = false;

    /**
     * Creates and registers a new MonacoModel instance.
     * @returns The reactive MonacoModel object.
     */
    public static createModel(
        id: string,
        initialContent: string,
        contentChangeListener: (
            e: monaco.editor.IModelContentChangedEvent,
        ) => void,
    ): MonacoModel {
        if (ModelManager.models.has(id)) {
            console.warn(
                `Model ID '${id}' already exists. Returning existing model.`,
            );
            return ModelManager.models.get(id)!;
        }

        const newModel = new MonacoModel(
            id,
            initialContent,
            contentChangeListener,
        );
        ModelManager.models.set(id, newModel);

        return newModel;
    }

    /**
     * Retrieves the reactive MonacoModel object for a given ID.
     */
    public static getModel(id: string): MonacoModel | undefined {
        return ModelManager.models.get(id);
    }

    /**
     * Disposes of a model.
     */
    public static disposeModel(id: string): void {
        const reactiveModel = ModelManager.models.get(id);
        if (reactiveModel) {
            reactiveModel.dispose(); // Dispose the Monaco model and listener
            ModelManager.models.delete(id);
            console.log(`Model '${id}' disposed.`);
        }
    }
}

export function Environment() {
    self.MonacoEnvironment = {
        getWorker() {
            return new tsWorker();
        },
    };
}

export async function Editor(
    domElement: HTMLElement,
    options: monaco.editor.IStandaloneEditorConstructionOptions,
) {
    return monaco.editor.create(domElement, options);
}
