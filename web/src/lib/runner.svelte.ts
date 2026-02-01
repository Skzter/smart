import { toast } from "svelte-sonner";
import { Mutex, timeout } from "async-ts";
import {
    getAuthHeaders,
    getMedia,
    getScreenshotUrl,
    getVideoUrl,
    runContainer,
    saveTestLocal,
} from "./api";
import { buildStepTree } from "$lib/runnerlogtransform";
import type { SaveState } from "$types/save";
import { baseURL } from "./shared.svelte";

export class Runner {
    private chatId: string;
    private userId: string;

    private m: Mutex = new Mutex();
    private running: boolean = $state(false);

    private storageState: SaveState = $state("idle");
    private storedTest: string = $state("");

    private streamAbortController?: AbortController;
    private retryCount = 0;
    private readonly MAX_RETRIES = 5;

    private fetchingMedia: boolean = false;
    private mediaPollTimeoutId: ReturnType<typeof setTimeout> | null = null;
    private mediaPollCount = 0;
    private static readonly MAX_MEDIA_POLL_ATTEMPTS = 12; // 12 * 5s = 60s
    private static readonly MEDIA_POLL_INTERVAL_MS = 5000;

    public videoUrl = $state<string | null>(null);
    public screenshotUrl = $state<string | null>(null);

    public logStatus = $state<
        "idle" | "connecting" | "connected" | "error" | "closed"
    >("idle");

    public logError = $state<string | null>(null);

    public result: { begin: string; end?: string }[] = $state([]);

    public model = $derived.by<Model>(() => {
        const r = this.result as RunnerResult;

        if (r && typeof r === "object" && "summary" in r && "steps" in r) {
            return r as Model;
        }

        if (Array.isArray(r)) {
            return buildStepTree(r);
        }

        return {
            summary: { status: "idle" },
            steps: [],
        };
    });

    constructor(chatId: string, userId: string, initialTestId = "") {
        this.chatId = chatId;
        this.userId = userId;
        this.storedTest = initialTestId;
    }

    public isRunning(): boolean {
        return this.running;
    }

    public async setTest(id: string) {
        if (this.running) {
            throw Error("Es läuft momentan ein Test");
        }
        const release = await this.m.obtain();
        this.storedTest = id;
        release();
    }

    public async storeTest(testcode: string) {
        if (!this.userId || !this.chatId) {
            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer- oder Konversations-ID fehlt.",
            });
            return;
        }

        const sanitizedUserId = this.userId.includes("|")
            ? this.userId.split("|")[1]
            : this.userId;

        this.storageState = "saving";

        try {
            const test = await saveTestLocal({
                userId: sanitizedUserId,
                chatId: this.chatId,
                code: testcode,
            });

            await this.setTest(test.testcaseId);
            this.storageState = "success";
            toast.success("Test erfolgreich gespeichert!");
        } catch (error) {
            this.storageState = "error";
            toast.error("Speichern fehlgeschlagen", {
                description:
                    error instanceof Error
                        ? error.message
                        : "Unbekannter Fehler",
            });
        } finally {
            setTimeout(() => {
                this.storageState = "idle";
            }, 2000);
        }
    }

    public getCurTest(): string {
        return this.storedTest;
    }

    public getStorageState(): SaveState {
        return this.storageState;
    }

    private startLogStream(testId: string) {
        this.logStatus = "connecting";
        this.logError = null;

        const connect = async () => {
            this.streamAbortController = new AbortController();
            const signal = this.streamAbortController.signal;

            try {
                const response = await fetch(
                    `${baseURL}/test/${testId}/stream`,
                    {
                        headers: {
                            Accept: "text/event-stream",
                            ...getAuthHeaders(),
                        },
                        signal,
                    },
                );

                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`);
                }

                this.retryCount = 0;
                this.logStatus = "connected";

                const reader = response.body?.getReader();
                const decoder = new TextDecoder();
                if (!reader) {
                    throw new Error("No response body");
                }

                const processLines = (lines: string[]) => {
                    let eventType = "log";
                    const dataLines: string[] = [];
                    for (const line of lines) {
                        if (line.startsWith("event:")) {
                            eventType = line.slice(6).trim();
                        } else if (line.startsWith("data:")) {
                            dataLines.push(line.slice(5));
                        } else if (line === "") {
                            const data = dataLines.join("\n").trim();
                            if (eventType === "log" && data) {
                                this.result = [...this.result, { begin: data }];
                            } else if (eventType === "finished") {
                                this.stopLogStream();
                                return true;
                            }
                            eventType = "log";
                            dataLines.length = 0;
                        }
                    }
                    return false;
                };

                let buffer = "";
                while (true) {
                    const { done, value } = await reader.read();
                    if (value) {
                        buffer += decoder.decode(value, { stream: true });
                    }
                    const lines = buffer.split("\n");
                    buffer = lines.pop() ?? "";
                    if (processLines(lines)) return;
                    if (done) {
                        if (buffer.trim() && processLines([...buffer.split("\n"), ""])) return;
                        break;
                    }
                }
                this.stopLogStream();
            } catch {
                if (signal.aborted) return;
                this.streamAbortController = undefined;
                if (this.retryCount < this.MAX_RETRIES) {
                    this.retryCount++;
                    const delay = 1000 * this.retryCount;
                    setTimeout(connect, delay);
                } else {
                    this.logStatus = "error";
                    this.logError = "Live-Log-Verbindung fehlgeschlagen";
                }
            }
        };

        connect();
    }

    private stopLogStream() {
        this.streamAbortController?.abort();
        this.streamAbortController = undefined;
        this.logStatus = "closed";
    }

    // =======================
    // TEST EXECUTION
    // =======================

    public async run() {
        this.result = [];

        if (!this.userId || !this.chatId || !this.storedTest) {
            console.error(
                `Missing IDs - ChatID: ${this.chatId}, UserID: ${this.userId}, TestID: ${this.storedTest}`,
            );

            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer-, -Konversations oder Test-ID fehlt.",
            });
            return;
        }

        const release = await this.m.obtain();
        if (this.running) {
            toast.error("Es läuft bereits ein Test", {
                description: `Id: ${this.storedTest}`,
            });
            release();
            return;
        }

        this.running = true;
        release();

        toast.message("Test wird ausgeführt", {
            description: `Id: ${this.storedTest}`,
        });

        const sanitizedUserId = this.userId.includes("|")
            ? this.userId.split("|")[1]
            : this.userId;

        try {
            await runContainer({
                userId: sanitizedUserId,
                testId: this.getCurTest(),
                chatId: this.chatId,
            });

            this.startLogStream(this.getCurTest());
        } catch (error) {
            console.error("Error running test:", error);
            this.logStatus = "error";
            this.running = false;
        }
    }

    public async fetchMediaUrl() {
        // Prevent multiple concurrent fetches
        if (this.fetchingMedia) {
            return;
        }
        this.fetchingMedia = true;

        try {
            for (let i = 0; i < this.MAX_RETRIES; i++) {
                await timeout(200 * (i + 1));
                try {
                    const resp = await getMedia(this.getCurTest());
                    const testId = this.getCurTest();
                    if (resp.hasVideo) {
                        this.videoUrl = await getVideoUrl(testId);
                    }
                    if (resp.hasScreenshot) {
                        this.screenshotUrl = await getScreenshotUrl(testId);
                    }
                    if (resp.hasVideo || resp.hasScreenshot) {
                        this.mediaPollCount = 0;
                        return;
                    }
                } catch (error) {
                    // Don't retry on errors - they likely won't resolve with retries
                    console.error("Error fetching media:", error);
                    break;
                }
            }
            // Video/screenshot not ready yet (upload in progress): poll again later
            if (
                this.mediaPollCount < Runner.MAX_MEDIA_POLL_ATTEMPTS &&
                this.model.summary.status === "failed"
            ) {
                this.mediaPollCount++;
                this.mediaPollTimeoutId = setTimeout(() => {
                    this.mediaPollTimeoutId = null;
                    this.fetchMediaUrl();
                }, Runner.MEDIA_POLL_INTERVAL_MS);
            } else {
                this.clearMediaUrls();
            }
        } finally {
            this.fetchingMedia = false;
        }
    }

    public clearVideoUrl() {
        this.clearMediaUrls();
    }

    private clearMediaUrls() {
        if (this.mediaPollTimeoutId != null) {
            clearTimeout(this.mediaPollTimeoutId);
            this.mediaPollTimeoutId = null;
        }
        this.mediaPollCount = 0;
        this.videoUrl = null;
        this.screenshotUrl = null;
    }
}

type Summary = {
    status: "idle" | "running" | "passed" | "failed";
    durationMs?: number;
};

type Step = {
    kind?: "group" | "step";
    label: string;
    status?: "running" | "done" | "failed";
    children?: Step[];
};

type Model = {
    summary: Summary;
    steps: Step[];
};

type RunnerResult = Model | { begin: string }[] | null | undefined;
