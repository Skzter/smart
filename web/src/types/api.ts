export type ApiMessage = {
    id: string;
    body: string;
    role: "user" | "assistant";
    createdAt: string;
    type: "Validation" | "Generation" | "Any";
};

export type ApiChatSummary = {
    chatId: string;
    userId: string;
    title: string;
    createdAt: string;
    updatedAt: string;
};

export type ApiChatRequest = {
    prompt: string;
    userId: string;
    chatId: string;
};

export type ApiChatResponse = {
    message: ApiMessage;
    userId: string;
    chatId: string;
    title?: string;
};

export type ApiSaveTestLocal = {
    userId: string;
    chatId: string;
    code: string;
};

export type ApiSaveTestLocalResponse = {
    testcaseId: string;
    action: string;
};

export type ApiRunContainer = {
    userId: string;
    testId: string;
    chatId: string;
};

export type ApiGetChatByIdResponse = {
    id: string;
    userId: string;
    createdAt: string;
    updatedAt: string;
    title: string;
    messages: ApiMessage[];
    lastTest: string;
    lastAutoPlaywrightPrompt: string;
};

export type ApiToken = {
    userId: string;
    token: string;
    createdAt: Date;
    updatedAt: Date;
    expiresAt: Date;
    revokedAt: Date | null;
};

export type ApiChatsRequest = {
    page: number;
    groupIds: string[];
};

export type ApiChatsResponse = {
    summaries: ApiChatSummary[];
    pageSize: number;
    hasMore: boolean;
};
