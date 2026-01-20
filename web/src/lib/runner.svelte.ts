import { toast } from "svelte-sonner";
import { Mutex } from "async-ts";
import { runContainer, saveTestLocal } from "./api";
import type { SaveState } from "$types/save";

export class Runner {
    private chatId: string;
    private userId: string;

    private m: Mutex = new Mutex();
    private running: boolean = $state(false);

    private storageState: SaveState = $state("idle");
    private storedTest: string = $state("");

    private eventSource?: EventSource;
    private retryCount = 0;
    private readonly MAX_RETRIES = 3;

    public logStatus = $state<
        "idle" | "connecting" | "connected" | "error" | "closed"
    >("idle");

    public logError = $state<string | null>(null);

    public result: { begin: string; end?: string }[] = $state([]);

    constructor(chatId: string, userId: string) {
        this.chatId = chatId;
        this.userId = userId;
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

        const connect = () => {
            this.eventSource = new EventSource(`/api/v1/test/${testId}/stream`);

            this.eventSource.onopen = () => {
                this.retryCount = 0;
                this.logStatus = "connected";
            };

            this.eventSource.addEventListener("log", (event) => {
                const e = event as MessageEvent<string>;
                this.result = [...this.result, { begin: e.data }];
            });

            this.eventSource.addEventListener("finished", () => {
                this.stopLogStream();
            });

            this.eventSource.onerror = () => {
                this.eventSource?.close();

                if (this.retryCount < this.MAX_RETRIES) {
                    this.retryCount++;
                    const delay = 1000 * this.retryCount;
                    setTimeout(connect, delay);
                } else {
                    this.logStatus = "error";
                    this.logError = "Live-Log-Verbindung fehlgeschlagen";
                }
            };
        };

        connect();
    }

    private stopLogStream() {
        this.eventSource?.close();
        this.eventSource = undefined;
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
}
