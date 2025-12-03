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
    createdAt: number;
    updatedAt: number;
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
