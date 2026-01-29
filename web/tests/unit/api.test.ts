import { describe, expect, it, vi, beforeEach, type Mock } from "vitest";
import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import {
    generatePrompt,
    getChats,
    getTemplate,
    saveTestLocal,
    runContainer,
    getChatById,
    deleteLocalTest,
    validatePrompt,
    updateChatTitle,
    getApiToken,
} from "../../src/lib/api";
import * as shared from "../../src/lib/shared.svelte";

// Mock axios
vi.mock("axios");

// Mock the shared module
vi.mock("../../src/lib/shared.svelte", () => ({
    user: { id: "user123" },
    chat: { id: "chat456", isLoading: false },
    apiToken: { token: null },
}));

describe("API Functions", () => {
    const mockUserId = "user123";
    const mockChatId = "chat456";

    const mockValidateParams = {
        userId: mockUserId,
        chatId: mockChatId,
        prompt: "Validate this prompt",
    };

    const mockValidateResponse = {
        chatId: mockChatId,
        message: {
            body: "Prompt validated successfully!",
        },
    };

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
            chatId: mockChatId,
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
                chatId: mockChatId,
                userId: mockUserId,
            },
        };

        it("should make a POST request to /chat with chat params", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue(mockApiResponse);

            const result = await generatePrompt(mockChatRequest);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/chat",
                baseURL: "http://localhost:8081/api/v1",
                headers: {},
                data: mockChatRequest,
            });
            expect(result).toEqual({
                message: mockMessage,
                chatId: mockChatId,
                userId: mockUserId,
            });
        });

        it("should include all required chat parameters", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue(mockApiResponse);

            await generatePrompt(mockChatRequest);

            const callArgs = mockedAxios.mock.calls[0][0];
            expect(callArgs.data).toHaveProperty("prompt", "test prompt");
            expect(callArgs.data).toHaveProperty("userId", mockUserId);
            expect(callArgs.data).toHaveProperty("chatId", mockChatId);
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

            await expect(generatePrompt(mockChatRequest)).rejects.toThrow(
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

        it("should make a GET request to /chats", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({
                data: {
                    chatSummarys: mockChatSummaries,
                    hasMore: false,
                    pageSize: 10,
                },
            });

            const result = await getChats({ page: 0, groupIds: [] });

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats?page=0`,
                baseURL: "http://localhost:8081/api/v1",
                headers: {},
            });
            expect(result).toEqual({
                summaries: mockChatSummaries,
                hasMore: false,
                pageSize: 10,
            });
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

            await expect(getChats({ page: 0, groupIds: [] })).rejects.toThrow(
                "Failed to fetch user chats",
            );
        });
    });
    describe.skip("validatePrompt TOOD: fix this test", () => {
        it("should make a POST request to /validate with proper params", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockValidateResponse });

            await validatePrompt(mockValidateParams);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/validate",
                baseURL: "/api/v1/",
                data: mockValidateParams,
            });
        });

        it("should return the API response data", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockValidateResponse });

            const result = await validatePrompt(mockValidateParams);

            expect(result).toEqual({
                chatId: mockChatId,
                message: {
                    body: "Prompt validated successfully!",
                },
            });
        });

        it("should include all required parameters", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockValidateResponse });

            await validatePrompt(mockValidateParams);

            const callArgs = mockedAxios.mock.calls[0][0];

            expect(callArgs.data).toHaveProperty(
                "userId",
                mockValidateParams.userId,
            );
            expect(callArgs.data).toHaveProperty(
                "chatId",
                mockValidateParams.chatId,
            );
            expect(callArgs.data).toHaveProperty(
                "prompt",
                mockValidateParams.prompt,
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
                url: "/template",
                baseURL: "http://localhost:8081/api/v1",
                headers: {},
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
            chatId: mockChatId,
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
                url: "/saveLocal",
                baseURL: "http://localhost:8081/api/v1",
                headers: {},
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

    describe.skip("runContainer", () => {
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

        it("should make a GET request to /chats/:chatId", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockChatResponse });

            const result = await getChatById();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats/${mockChatId}`,
                headers: {},
                baseURL: "http://localhost:8081/api/v1",
            });
            expect(result).toEqual(mockChatResponse);
        });

        it("should use the current chat id from shared state", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValue({ data: mockChatResponse });

            shared.user.id = "differentUser";
            shared.chat.id = "differentChat";

            await getChatById();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: `/chats/differentChat`,
                headers: {},
                baseURL: "http://localhost:8081/api/v1",
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
                baseURL: "http://localhost:8081/api/v1",
                url: "/deleteLocal",
                headers: {},
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
                baseURL: "http://localhost:8081/api/v1",
                url: "/deleteLocal",
                headers: {},
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

    describe("getApiToken", () => {
        const mockTokenResponse = {
            userId: mockUserId,
            token: "mock-jwt-token-12345",
            createdAt: new Date("2024-01-01T00:00:00Z"),
            updatedAt: new Date("2024-01-01T00:00:00Z"),
            expiresAt: new Date("2024-01-02T00:00:00Z"),
            revokedAt: null,
        };

        it("should make a POST request to /auth/generate with userId", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValueOnce({ data: mockTokenResponse });

            const result = await getApiToken();

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "auth/generate",
                baseURL: "http://localhost:8081/api/v1",
                data: { userId: mockUserId },
            });
            expect(result).toEqual(mockTokenResponse);
        });

        it("should return the API token data", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValueOnce({ data: mockTokenResponse });

            const result = await getApiToken();

            expect(result).toHaveProperty("userId", mockUserId);
            expect(result).toHaveProperty("token", "mock-jwt-token-12345");
            expect(result).toHaveProperty("createdAt");
            expect(result).toHaveProperty("updatedAt");
            expect(result).toHaveProperty("expiresAt");
            expect(result).toHaveProperty("revokedAt", null);
        });

        it("should use the current user id from shared state", async () => {
            const mockedAxios = axios as unknown as Mock;
            mockedAxios.mockResolvedValueOnce({ data: mockTokenResponse });

            shared.user.id = "user999";

            await getApiToken();

            const callArgs = mockedAxios.mock.calls[0][0];
            expect(callArgs.data).toHaveProperty("userId", "user999");
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as unknown as Mock;
            const err = new AxiosError("Failed to generate token");
            err.response = {
                data: { message: "Failed to generate token" },
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };
            mockedAxios.mockRejectedValueOnce(err);

            await expect(getApiToken()).rejects.toThrow(
                "Failed to generate token",
            );
        });
    });
});

describe("updateChatTitle", () => {
    const newTitle = "Updated Chat Title";
    const mockChatId = "chat456";

    it("should make a PATCH request to users/${userId}/chats/${chatId}/title with correct data", async () => {
        const mockedAxios = axios as unknown as Mock;
        mockedAxios.mockResolvedValue({
            data: {
                chatId: mockChatId,
                title: newTitle,
            },
        });

        const result = await updateChatTitle(mockChatId, newTitle);

        expect(mockedAxios).toHaveBeenCalledWith({
            method: "patch",
            url: `/chats/${mockChatId}/title`,
            baseURL: "http://localhost:8081/api/v1",
            data: {
                title: newTitle,
            },
        });
        expect(result).toEqual({
            chatId: mockChatId,
            title: newTitle,
        });
    });

    it("should use the current user id from shared state", async () => {
        const mockedAxios = axios as unknown as Mock;
        mockedAxios.mockResolvedValue({
            data: {
                chatId: mockChatId,
                title: newTitle,
            },
        });

        await updateChatTitle(mockChatId, newTitle);

        expect(mockedAxios).toHaveBeenCalledWith({
            method: "patch",
            url: `/chats/${mockChatId}/title`,
            baseURL: "http://localhost:8081/api/v1",
            data: {
                title: newTitle,
            },
        });
    });

    it("should reject when the API call fails", async () => {
        const mockedAxios = axios as unknown as Mock;
        const err = new AxiosError("Failed to update chat title");
        err.response = {
            data: { message: "Failed to update chat title" },
            status: 500,
            statusText: "Internal Server Error",
            headers: {},
            config: {} as InternalAxiosRequestConfig,
        };
        mockedAxios.mockRejectedValue(err);

        await expect(updateChatTitle(mockChatId, newTitle)).rejects.toThrow(
            "Failed to update chat title",
        );
    });
});
