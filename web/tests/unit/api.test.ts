import { describe, expect, it, vi, beforeEach, type Mock } from "vitest";
import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import {
    getChatResponse,
    getUserChats,
    getTemplate,
    saveTestLocal,
    runContainer,
    getChatById,
    deleteLocalTest,
} from "../../src/lib/api";
import * as shared from "../../src/lib/shared.svelte";

// Mock axios
vi.mock("axios");

// Mock the shared module
vi.mock("../../src/lib/shared.svelte", () => ({
    user: { id: "user123" },
    chat: { id: "chat456", isLoading: false },
}));

describe("API Functions", () => {
    const mockUserId = "user123";
    const mockConversationId = "conv456";
    const mockChatId = "chat456";

    beforeEach(() => {
        vi.clearAllMocks();
        // Reset mock values
        shared.user.id = mockUserId;
        shared.chat.id = mockChatId;
    });

    describe("getChatResponse", () => {
        const mockChatRequest = {
            prompt: "test prompt",
            userId: mockUserId,
            conversationId: mockConversationId,
        };

        const mockMessage = {
            id: "msg123",
            body: "test response",
            role: "assistant",
            createdAt: "2024-01-01T00:00:00Z",
        };

        const mockApiResponse = {
            data: {
                message: mockMessage,
                conversationId: mockConversationId,
                userId: mockUserId,
            },
        };

        it("should make a POST request to /chat with chat params", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue(mockApiResponse);

            const result = await getChatResponse(mockChatRequest);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "chat",
                baseURL: "http://localhost:8081/api/v1/",
                data: mockChatRequest,
            });
            expect(result).toEqual({
                message: mockMessage,
                conversationId: mockConversationId,
                userId: mockUserId,
            });
        });

        it("should include all required chat parameters", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue(mockApiResponse);

            await getChatResponse(mockChatRequest);

            const callArgs = mockedAxios.mock.calls[0][0];
            expect(callArgs.data).toHaveProperty("prompt", "test prompt");
            expect(callArgs.data).toHaveProperty("userId", mockUserId);
            expect(callArgs.data).toHaveProperty(
                "conversationId",
                mockConversationId,
            );
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Chat service unavailable");
            err.response = {
                data: { message: "Chat service unavailable" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(getChatResponse(mockChatRequest)).rejects.toThrow(
                "Chat service unavailable",
            );
        });
    });

    describe("getUserChats", () => {
        const mockChatSummaries = [
            {
                chatId: "chat1",
                userId: mockUserId,
                title: "Test Chat 1",
                createdAt: "2024-01-01T00:00:00Z",
                updatedAt: "2024-01-01T00:00:00Z",
            },
            {
                chatId: "chat2",
                userId: mockUserId,
                title: "Test Chat 2",
                createdAt: "2024-01-02T00:00:00Z",
                updatedAt: "2024-01-02T00:00:00Z",
            },
        ];

        it("should make a GET request to /chats/:userId", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({
                data: { chatSummarys: mockChatSummaries },
            });

            const result = await getChats();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats/${mockUserId}`,
                baseURL: "http://localhost:8081/api/v1/",
            });
            expect(result).toEqual(mockChatSummaries);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Failed to fetch user chats");
            err.response = {
                data: { message: "Failed to fetch user chats" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(getChats()).rejects.toThrow(
                "Failed to fetch user chats",
            );
        });
    });

    describe("getTemplate", () => {
        const mockTemplate = "test template content";

        it("should make a GET request to /template and return data", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({
                data: { template: mockTemplate },
            });

            const result = await getTemplate();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: "template",
                baseURL: "http://localhost:8081/api/v1/",
            });
            expect(result).toEqual(mockTemplate);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Template not found");
            err.response = {
                data: { message: "Template not found" },
                status: 404,
                statusText: "Not Found",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(getTemplate()).rejects.toThrow("Template not found");
        });
    });

    describe("saveTestLocal", () => {
        const mockSaveLocalRequest = {
            code: "fantastic code",
            userId: mockUserId,
            conversationId: mockConversationId,
        };

        const mockSaveLocalResponse = {
            testcaseId: "testid",
            action: "saved",
        };

        it("should make a POST request to /saveLocal and return data", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockSaveLocalResponse });

            const result = await saveTestLocal(mockSaveLocalRequest);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "saveLocal",
                baseURL: "http://localhost:8081/api/v1/",
                data: mockSaveLocalRequest,
            });
            expect(result).toEqual(mockSaveLocalResponse);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Failed to save test");
            err.response = {
                data: { message: "Failed to save test" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(saveTestLocal(mockSaveLocalRequest)).rejects.toThrow(
                "Failed to save test",
            );
        });
    });

    describe("runContainer", () => {
        const mockParams = {
            userId: mockUserId,
            testId: "test123",
            chatId: "chat456",
        };
        const mockResult = "Container executed successfully";

        it("should make a POST request to /run and return data", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: { result: mockResult } });

            const result = await runContainer(mockParams);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "run",
                baseURL: "http://localhost:8081/api/v1/",
                data: mockParams,
            });
            expect(result).toEqual(mockResult);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Failed to run container");
            err.response = {
                data: { message: "Failed to run container" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(runContainer(mockParams)).rejects.toThrow(
                "Failed to run container",
            );
        });

        it("should pass the correct params as request body", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: { result: mockResult } });
            const complexParams = {
                userId: "user999",
                testId: "test999",
                chatId: "chat999",
            };

            await runContainer(complexParams);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "run",
                baseURL: "http://localhost:8081/api/v1/",
                data: complexParams,
            });
        });
    });

    describe("getChatById", () => {
        const mockChatResponse = {
            id: mockChatId,
            userId: mockUserId,
            createdAt: "2024-01-01T00:00:00Z",
            updatedAt: "2024-01-01T00:00:00Z",
            title: "Test Chat",
            messages: [
                {
                    id: "msg1",
                    body: "Hello",
                    role: "user",
                    createdAt: "2024-01-01T00:00:00Z",
                },
                {
                    id: "msg2",
                    body: "Hi there",
                    role: "assistant",
                    createdAt: "2024-01-01T00:00:01Z",
                },
            ],
            lastTest: "test code here",
            lastAutoPlaywrightPrompt: "last prompt here",
        };

        it("should make a GET request to /users/:userId/chats/:chatId", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockChatResponse });

            const result = await getChatById();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats/${mockChatId}`,
                baseURL: "http://localhost:8081/api/v1/",
            });
            expect(result).toEqual(mockChatResponse);
        });

        it("should use the current user and chat ids from shared state", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockChatResponse });

            shared.user.id = "differentUser";
            shared.chat.id = "differentChat";

            await getChatById();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats/differentChat`,
                baseURL: "http://localhost:8081/api/v1/",
            });
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Chat not found");
            err.response = {
                data: { message: "Chat not found" },
                status: 404,
                statusText: "Not Found",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(getChatById()).rejects.toThrow("Chat not found");
        });
    });

    describe("deleteLocalTest", () => {
        const mockTestcaseId = "test123";

        it("should make a DELETE request to /deleteLocal with correct params", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({
                data: "Test deleted successfully",
            });

            const result = await deleteLocalTest(mockTestcaseId);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "delete",
                baseURL: "http://localhost:8081/api/v1/",
                url: "/deleteLocal",
                params: {
                    testcaseId: mockTestcaseId,
                    chatId: mockChatId,
                    userId: mockUserId,
                },
            });
            expect(result).toEqual("Test deleted successfully");
        });

        it("should use the current user and chat ids from shared state", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: "Test deleted" });

            shared.user.id = "user999";
            shared.chat.id = "chat999";

            await deleteLocalTest(mockTestcaseId);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "delete",
                baseURL: "http://localhost:8081/api/v1/",
                url: "/deleteLocal",
                params: {
                    testcaseId: mockTestcaseId,
                    chatId: "chat999",
                    userId: "user999",
                },
            });
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Failed to delete test");
            err.response = {
                data: { message: "Failed to delete test" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValue(err);

            await expect(deleteLocalTest(mockTestcaseId)).rejects.toThrow(
                "Failed to delete test",
            );
        });
    });
});
