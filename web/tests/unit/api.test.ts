import { describe, expect, it, vi, beforeEach } from "vitest";
import axios from "axios";
import {
    getChatResponse,
    getUserInfo,
    getTemplate,
} from "../../src/lib/Api.ts";

// Mock axios
vi.mock("axios");

describe("API Functions", () => {
    const mockUserId = "user123";
    const mockConversationId = "conv456";
    const mockMessage = { data: "test message", agent: "user" };

    const mockUserParams = {
        userId: mockUserId,
    };

    const mockChatParams = {
        message: mockMessage,
        userId: mockUserId,
        conversationId: mockConversationId,
    };

    const mockResponseData = { data: "test data" };

    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe("getUserInfo", () => {
        it("should make a POST request to /userInfo with user params", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue({ data: mockResponseData });

            const result = await getUserInfo(mockUserParams, "/userInfo");

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/userInfo",
                baseURL: "/api/v1/chat",
                data: mockUserParams,
            });
            expect(result.data).toEqual(mockResponseData);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockRejectedValue(
                new Error("Failed to fetch user info"),
            );

            await expect(
                getUserInfo(mockUserParams, "/userInfo"),
            ).rejects.toThrow("Failed to fetch user info");
        });
    });

    describe("getChatResponse", () => {
        it("should make a POST request to /chat with chat params", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue({ data: mockResponseData });

            const result = await getChatResponse(mockChatParams, "/chat");

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/chat",
                baseURL: "/api/v1/chat",
                data: mockChatParams,
            });
            expect(result.data).toEqual(mockResponseData);
        });

        it("should include all required chat parameters", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue({ data: mockResponseData });

            await getChatResponse(mockChatParams, "/chat");

            const callArgs = mockedAxios.mock.calls[0][0];
            expect(callArgs.data).toHaveProperty("message");
            expect(callArgs.data.message).toEqual(mockMessage);
            expect(callArgs.data).toHaveProperty("userId", mockUserId);
            expect(callArgs.data).toHaveProperty(
                "conversationId",
                mockConversationId,
            );
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockRejectedValue(
                new Error("Chat service unavailable"),
            );

            await expect(
                getChatResponse(mockChatParams, "/chat"),
            ).rejects.toThrow("Chat service unavailable");
        });
    });

    describe("getTemplate", () => {
        it("should make a GET request to /template and return data", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue({ data: mockResponseData });

            const result = await getTemplate("/template");

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "get",
                url: "/template",
                baseURL: "/api/v1/chat",
            });
            expect(result.data).toEqual(mockResponseData);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockRejectedValue(new Error("Template not found"));

            await expect(getTemplate("/template")).rejects.toThrow(
                "Template not found",
            );
        });
    });
});
