import type { Message } from "$types/message";
import { toast } from "svelte-sonner";
import { Mutex } from "async-ts";

export const messages: Message[] = $state([]);

export const user = $state({
    id: "",
});

export const chat = $state({
    id: "",
    isLoading: false,
});

class Runner {
    private m: Mutex = new Mutex();
    private running: boolean = $state(false);
    private currTestId: string = "";

    public result: string = $state("");

    public isRunning(): boolean {
        return this.running;
    }

    public async setTest(id: string) {
        if (this.running) {
            toast.error(
                "Test konnte nicht gespeichert werden, es läuft bereits einer",
                {
                    description: `Id: ${this.currTestId}`,
                },
            );
            return;
        } else {
            const release = await this.m.obtain();
            this.currTestId = id;
            release();
        }
    }

    public getCurTest(): string {
        return this.currTestId;
    }

    public async run() {
        if (!user.id || !chat.id || !this.currTestId) {
            console.error(
                "Missing IDs - ChatID: " +
                    chat.id +
                    " UserID: " +
                    user.id +
                    " TestID: " +
                    this.currTestId,
            );
            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer-, -Konversations oder Test-ID fehlt.",
            });
            return;
        }

        const release = await this.m.obtain();
        if (this.running) {
            toast.error("Es läuft bereits ein Test", {
                description: `Id: ${this.currTestId}`,
            });
            release();
            return;
        } else {
            this.running = true;
            release();
        }

        toast.message("Test wird ausgeführt", {
            description: `Id: ${this.currTestId}`,
        });

        const sleep = (ms: number) => {
            return new Promise((resolve) => setTimeout(resolve, ms));
        };

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
        } finally {
            this.running = false;
        }
        */

        toast.message("Testausführung beendet");
    }
}

export const runner = $state(new Runner());
