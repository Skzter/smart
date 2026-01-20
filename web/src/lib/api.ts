import axios, { AxiosError } from "axios";
import type {
    ApiMessage,
    ApiChatSummary,
    ApiChatRequest,
    ApiChatResponse,
    ApiSaveTestLocal,
    ApiSaveTestLocalResponse,
    ApiRunContainer,
    ApiGetChatByIdResponse,
    ApiGroup,
    ApiCreateGroupRequest,
    ApiCreateGroupResponse,
} from "$types/api";
import { chat, user } from "./shared.svelte";

const baseURL = "http://localhost:8081/api/v1/";

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
            url: "chat",
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
            baseURL: "/api/v1/",
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
export async function getChats(): Promise<ApiChatSummary[]> {
    try {
        const response = await axios({
            method: "get",
            url: `/chats`,
            baseURL: baseURL,
        });
        return response.data.chatSummarys;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function getTemplate(): Promise<string> {
    try {
        const response = await axios({
            method: "get",
            url: "template",
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
            url: "saveLocal",
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
export async function runContainer(request: ApiRunContainer): Promise<void> {
    await axios({
        method: "post",
        url: "run",
        baseURL,
        data: request,
    });
}

export async function getChatById(chatId?: string, ): Promise<ApiGetChatByIdResponse> {
    try {
        const id = chatId ?? chat.id;
        if (!id) {
            throw new Error("chatId is required");
        }

        const response = await axios({
            method: "get",
            url: `/chats/${id}`,
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
  
export async function getGroups(): Promise<ApiGroup[]> {
    try {
        const response = await axios({ method: "get", url: "/groups", baseURL });
        return response.data as ApiGroup[];
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function createGroup(request: ApiCreateGroupRequest,): Promise<ApiCreateGroupResponse> {
    try {
        const response = await axios({ method: "post", url: "/groups", baseURL, data: request });
        return response.data as ApiCreateGroupResponse;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function assignChatToGroups(chatId: string, groupIds: string[]): Promise<void> {
    try {
        await axios({
        method: "post",
        url: `/chats/${chatId}/groups`,
        baseURL,
        data: { groupIds }, 
        });
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function removeChatFromGroup(chatId: string, groupId: string): Promise<void> {
    try {
        await axios({
        method: "delete",
        url: `/chats/${chatId}/groups/${groupId}`,
        baseURL,
        });
    } catch (error) {
        throw getErrorMessage(error);
    }
}  
