import type { DateRange } from "bits-ui";

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
