import axios, { AxiosError } from "axios";
import type {
    ApiMessage,
    ApiChatRequest,
    ApiChatResponse,
    ApiSaveTestLocal,
    ApiSaveTestLocalResponse,
    ApiRunContainer,
    ApiGetChatByIdResponse,
    ApiChatsRequest,
    ApiChatsResponse,
} from "$types/api";
import { chat, user } from "./shared.svelte";

const baseURL = "http://localhost:8081/api/v1";

/**
 * Fetches data from the api and returns the data for the chat
 * @param params: parameters for api
 * @param url: url for api
 */
export async function generatePrompt(
    request: ApiChatRequest,
): Promise<ApiChatResponse> {
    try {
        const response = await axios({
            method: "post",
            url: "/chat",
            baseURL: baseURL,
            data: request,
        });
        return {
            message: response.data.message as ApiMessage,
            chatId: response.data.chatId,
            userId: response.data.userId,
        };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

/** Validates the prompt by sending it to the /validationRes endpoint
 * @param body: object containing userId, conversationId, and prompt
 */
export async function validatePrompt(
    request: ApiChatRequest,
): Promise<ApiChatResponse> {
    try {
        const response = await axios({
            method: "post",
            url: "/validate",
            baseURL: baseURL,
            data: request,
        });
        return {
            message: response.data.message as ApiMessage,
            chatId: response.data.chatId,
            userId: response.data.userId,
        };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

/**
 * Fetches data from the api and returns the data for the user
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getChats(
    request: ApiChatsRequest,
): Promise<ApiChatsResponse> {
    const groups =
        request.groupIds.length > 0
            ? `&groups=${request.groupIds.join(",")}`
            : "";
    try {
        const response = await axios({
            method: "get",
            url: `/chats?page=${request.page}${groups}`,
            baseURL: baseURL,
        });
        return {
            summaries: response.data.chatSummarys,
            hasMore: response.data.hasMore,
            pageSize: response.data.pageSize,
        };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function getTemplate(): Promise<string> {
    try {
        const response = await axios({
            method: "get",
            url: "/template",
            baseURL: baseURL,
        });
        return response.data.template;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function saveTestLocal(
    request: ApiSaveTestLocal,
): Promise<ApiSaveTestLocalResponse> {
    try {
        const response = await axios({
            method: "post",
            url: "/saveLocal",
            baseURL: baseURL,
            data: request,
        });
        return {
            testcaseId: response.data.testcaseId,
            action: response.data.action,
        };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

// must block exectution until test is done
export async function runContainer(
    request: ApiRunContainer,
    handler: {
        onStepEnd?: (message: string) => void;
        onError?: (error: Error) => void;
        onStepBegin?: (message: string) => void;
    },
): Promise<void> {
    try {
        handler.onStepBegin?.("Getting Results...");
        const response = await axios({
            method: "post",
            url: "/run",
            baseURL: baseURL,
            data: request,
        });
        handler.onStepEnd?.(response.data.result);
    } catch (error) {
        handler.onError?.(getErrorMessage(error));
    }
}

export async function getChatById(): Promise<ApiGetChatByIdResponse> {
    try {
        const response = await axios({
            method: "get",
            url: `/chats/${chat.id}`,
            baseURL: baseURL,
        });
        return response.data;
    } catch (err) {
        throw getErrorMessage(err);
    }
}

export async function deleteLocalTest(testcaseId: string): Promise<string> {
    try {
        const response = await axios({
            method: "delete",
            baseURL: baseURL,
            url: "/deleteLocal",
            params: {
                testcaseId,
                conversationId: chat.id,
                userId: user.id,
            },
        });
        return response.data;
    } catch (err) {
        throw getErrorMessage(err);
    }
}

function getErrorMessage(error: unknown): Error {
    if (error instanceof AxiosError) {
        return new Error(error.response?.data.message ?? error.message);
    } else {
        return new Error(
            "no axios error returned - something went horribly wrong",
        );
    }
}
