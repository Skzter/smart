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
