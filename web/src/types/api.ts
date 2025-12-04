export type ApiMessage = {
    id: string;
    body: string;
    role: string;
    createdAt: number;
};

export type ApiChatSummary = {
    chatId: string;
    userId: string;
    title: string;
    createdAt: string;
    updatedAt: string;
};

export type ApiChatRequest = {
    message: {
        body: string;
        role: string;
    };
    userId: string;
    conversationId: string;
};

export type ApiChatResponse = {
    message: ApiMessage;
    userId: string;
    conversationId: string;
};

export type ApiSaveTestLocal = {
    userId: string;
    conversationId: string;
    code: string;
};

export type ApiSaveTEstLocalResponse = {
    testcaseId: string;
    action: string;
};

export type ApiRunContainer = {
    userId: string;
    testId: string;
    sessionId: string;
};
