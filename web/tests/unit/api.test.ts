import { describe, expect, it, vi, beforeEach } from "vitest";
import axios from "axios";
import {
    getChatResponse,
    getUserInfo,
    getTemplate,
    saveTestLocal,
    runContainer,
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

    const mockSaveLocalRequest = {
        testcode: "fantatsic code",
        userId: mockUserId,
        conversationId: mockConversationId,
    };

    const mockSaveLocalResponse = {
        testcaseId: "testid",
        action: "saved",
    };

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
                baseURL: "/api/v1/",
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
                baseURL: "/api/v1/",
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
                baseURL: "/api/v1/",
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

    describe("saveTestLocal", () => {
        it("should make a POST request to /saveLocal and return data", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue({ data: mockSaveLocalResponse });

            const result = await saveTestLocal(mockSaveLocalRequest);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/saveLocal",
                baseURL: "/api/v1/",
                data: mockSaveLocalRequest,
            });
            expect(result.data).toEqual(mockSaveLocalResponse);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockRejectedValue(new Error("Failed to save test"));

            await expect(saveTestLocal(mockSaveLocalRequest)).rejects.toThrow(
                "Failed to save test",
            );
        });
    });

    describe("runContainer", () => {
        const mockParams = { image: "node:latest", command: "echo hello" };
        const mockResponse = {
            data: { containerId: "123", status: "running" },
        };

        it("should make a POST request to /run and return data", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue(mockResponse);

            const result = await runContainer(mockParams);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/run",
                baseURL: "/api/v1/",
                data: mockParams,
            });
            expect(result.data).toEqual(mockResponse.data);
        });

        it("should use custom URL when provided", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue(mockResponse);
            const customUrl = "/custom-run";

            const result = await runContainer(mockParams, customUrl);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: customUrl,
                baseURL: "/api/v1/",
                data: mockParams,
            });
            expect(result.data).toEqual(mockResponse.data);
        });

        it("should reject when the API call fails", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockRejectedValue(new Error("Failed to run container"));

            await expect(runContainer(mockParams)).rejects.toThrow(
                "Failed to run container",
            );
        });

        it("should pass the correct params as request body", async () => {
            const mockedAxios = axios as vi.Mocked<typeof axios>;
            mockedAxios.mockResolvedValue(mockResponse);
            const complexParams = {
                image: "python:3.9",
                command: "python script.py",
                env: { KEY: "value" },
                resources: { cpu: 2, memory: "1GB" },
            };

            const result = await runContainer(complexParams);

            expect(mockedAxios).toHaveBeenCalledWith({
                method: "post",
                url: "/run",
                baseURL: "/api/v1/",
                data: complexParams,
            });
            expect(result.data).toEqual(mockResponse.data);
        });
    });
});
