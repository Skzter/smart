import axios, { AxiosError } from "axios";
import type { ApiMessage, ApiChatSummary } from "../types/api";

const baseURL = "http://localhost:8081/api/v1/";

/**
 * Fetches data from the api and returns the data for the chat
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getChatResponse(request: {
    prompt: string;
    userId: string;
    conversationId: string;
}): Promise<{ message: ApiMessage; userId: string; conversationId: string }> {
    try {
        const response = await axios({
            method: "post",
            url: "chat",
            baseURL: baseURL,
            data: request,
        });
        return {
            message: response.data.message as ApiMessage,
            conversationId: response.data.conversationId,
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
 */ // schön mit Rechtschreibfehler vom main
export async function getUserChats(request: {
    userId: string;
}): Promise<{ chatSummarys: ApiChatSummary[] }> {
    try {
        const response = await axios({
            method: "post",
            url: `chats/${request.userId}`,
            baseURL: baseURL,
        });
        return { chatSummarys: response.data.chatSummarys };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function getTemplate(): Promise<{ template: string }> {
    try {
        const response = await axios({
            method: "get",
            url: "template",
            baseURL: baseURL,
        });
        return { template: response.data.template };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function saveTestLocal(request: {
    userId: string;
    conversationId: string;
    code: string;
}): Promise<{ testcaseId: string; action: string }> {
    try {
        const response = await axios({
            method: "post",
            url: "savelocal",
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

export async function runContainer(request: {
    userId: string;
    testId: string;
    sessionId: string;
}): Promise<{ result: string }> {
    try {
        const response = await axios({
            method: "post",
            url: "run",
            baseURL: baseURL,
            data: request,
        });
        return {
            result: response.data.result,
        };
    } catch (error) {
        throw getErrorMessage(error);
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
