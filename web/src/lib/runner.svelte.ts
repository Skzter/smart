import { toast } from "svelte-sonner";
import { Mutex } from "async-ts";
import { user, chat } from "$lib/shared.svelte";
import { saveTestLocal } from "./api";
import type { SaveState } from "$types/save";

function sleep(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}
export class Runner {
    private m: Mutex = new Mutex();
    private running: boolean = $state(false);

    private storageState: SaveState = $state("idle");
    private storedTest: string = $state("");

    public result: string = $state("");

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
        if (!user.id || !chat.id) {
            console.error(
                "Missing IDs - ChatID: " + chat.id + " UserID: " + user.id,
            );
            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer- oder Konversations-ID fehlt.",
            });
            return;
        }

        const sanitizedUserId = user.id.includes("|")
            ? user.id.split("|")[1]
            : user.id;

        this.storageState = "saving";

        try {
            const test = await saveTestLocal({
                userId: sanitizedUserId,
                conversationId: chat.id,
                code: testcode,
            });
            await sleep(3000);

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

    public getStorageState(): string {
        return this.storageState;
    }

    public async run() {
        if (!user.id || !chat.id || !this.storedTest) {
            console.error(
                "Missing IDs - ChatID: " +
                    chat.id +
                    " UserID: " +
                    user.id +
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
        }

        this.running = true;
        release();

        toast.message("Test wird ausgeführt", {
            class: "!bg-purple",
            style: "!bg-red",
            description: `Id: ${this.storedTest}`,
        });

        await sleep(5000);

        this.result = "asaskjjsdkakjdashdjaskjdhaskjdhjaskd";
        this.running = false;

        /*
        const sanitizedUserId = user.id.includes("|")
            ? user.id.split("|")[1]
            : user.id;

        
        try {
            const response = await runContainer({
                userId: sanitizedUserId,
                testId: this.currTestId,
                sessionId: chat.id,
            });
            this.result = response;
        } catch (error) {
            console.error("Error running test:", error);
            this.logStatus = "error";
            this.running = false;
        }
        */

        toast.message("Testausführung beendet");
    }
}
