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
        } else {
            const release = await this.m.obtain();
            this.storedTest = id;
            release();
        }
    }

    public async storeTest(testcode: string) {
        if (!this.userId || !this.chatId) {
            console.error(
                "Missing IDs - ChatID: " +
                    this.chatId +
                    " UserID: " +
                    this.userId,
            );
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
                conversationId: this.chatId,
                code: testcode,
            });

            this.setTest(test.testcaseId);
            this.storageState = "success";

            toast.success("Test erfolgreich gespeichert!");
        } catch (error) {
            this.storageState = "error";

            let errorMsg = "Unbekannter Fehler";
            if (error instanceof Error) {
                errorMsg = error.message;
            }

            toast.error("Speichern fehlgeschlagen", {
                description: errorMsg,
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

    public getStorageState(): string {
        return this.storageState;
    }

    public async run() {
        this.result = [];
        if (!this.userId || !this.chatId || !this.storedTest) {
            console.error(
                "Missing IDs - ChatID: " +
                    this.chatId +
                    " UserID: " +
                    this.userId +
                    " TestID: " +
                    this.storedTest,
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
        } else {
            this.running = true;
            release();
        }

        toast.message("Test wird ausgeführt", {
            class: "!bg-purple",
            style: "!bg-red",
            description: `Id: ${this.storedTest}`,
        });

        const sanitizedUserId = this.userId.includes("|")
            ? this.userId.split("|")[1]
            : this.userId;

        try {
            await runContainer(
                {
                    userId: sanitizedUserId,
                    testId: this.getCurTest(),
                    sessionId: this.chatId,
                },
                {
                    onStepBegin: (message) => {
                        this.result.push({ begin: message });
                    },
                    onStepEnd: (message) => {
                        this.result[this.result.length - 1].end = message;
                    },
                },
            );
        } catch (error) {
            console.error("Error running test:", error);
        } finally {
            this.running = false;
        }

        toast.message("Testausführung beendet");
    }
}
