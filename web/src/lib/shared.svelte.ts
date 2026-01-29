import type { DateRange } from "bits-ui";
import { getApiToken } from "$lib/api";
import { toast } from "svelte-sonner";
import { AxiosError } from "axios";
import type { ApiToken } from "$types/api";

let updater: ((chatId: string, title: string) => void) | null = null;

export const baseURL = "http://localhost:8081/api/v1";

export type MessageType = "user" | "validation" | "generation" | "error";
export type Message = { t: MessageType; Message: string };

export const messages: Message[] = $state([]);

export const user = $state({
    id: "" as string | undefined,
});

class ApiTokenStore {
    private _token = $state<string | null>(null);

    constructor() {
        if (typeof window !== "undefined") {
            this._token = localStorage.getItem("apiToken");
        }
    }

    get token(): string | null {
        return this._token;
    }

    set token(value: string | null) {
        this._token = value;
        if (typeof window !== "undefined") {
            if (value) {
                localStorage.setItem("apiToken", value);
            } else {
                localStorage.removeItem("apiToken");
            }
        }
    }
}

export const apiToken = new ApiTokenStore();

export const chat = $state({
    id: "",
    isLoading: false,
});

export const ChatDate = $state({
    Range: undefined as DateRange | undefined,
});

export const ChatFilter = $state({
    sortBy: "recent" as "recent" | "created",
    timeFilter: "all" as "all" | "today" | "week" | "month",
});

export async function getToken() {
    try {
        const token = (await getApiToken()) as ApiToken;
        apiToken.token = token.token;
    } catch (err) {
        if (err instanceof AxiosError) {
            const error = err.message;
            toast.error(error, {
                description: "Das war wohl nichts mit der Historie.",
            });
        }
        apiToken.token = null;
    }
}

export function registerChatTitleUpdater(
    fn: (chatId: string, title: string) => void,
) {
    updater = fn;
}

export function updateChatTitle(chatId: string, title: string) {
    updater?.(chatId, title);
}
