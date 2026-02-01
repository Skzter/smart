import axios, { AxiosError } from "axios";
import type {
    ApiMessage,
    ApiChatRequest,
    ApiChatResponse,
    ApiSaveTestLocal,
    ApiSaveTestLocalResponse,
    ApiRunContainer,
    ApiGetChatByIdResponse,
    ApiToken,
    ApiChatsRequest,
    ApiChatsResponse,
    ApiChatSummary,
    ApiMediaResponse,
    ApiGroup,
    ApiCreateGroupRequest,
    ApiCreateGroupResponse,
} from "$types/api";
import { chat, user, apiToken, baseURL } from "./shared.svelte";

export function getAuthHeaders(): Record<string, string> {
    return apiToken.token ? { Authorization: `Bearer ${apiToken.token}` } : {};
}

// Axios response interceptor to handle 401 and refresh token
axios.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;

        // Check if error is 401 and we haven't already retried
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;

            // Don't retry if the failed request was getApiToken itself
            if (originalRequest.url?.includes("auth/generate")) {
                return Promise.reject(error);
            }

            // If we don't have a user id (e.g. auth not ready / logged out), we can't refresh a token
            if (!user.id) {
                apiToken.token = null;
                return Promise.reject(error);
            }

            try {
                // Get new token
                const tokenResponse = await axios({
                    method: "post",
                    url: "auth/generate",
                    baseURL: baseURL,
                    data: { userId: user.id },
                });

                const newToken = tokenResponse.data as ApiToken;
                apiToken.token = newToken.token;

                // Retry original request with new token
                originalRequest.headers.Authorization = `Bearer ${newToken.token}`;
                return axios(originalRequest);
            } catch (refreshError) {
                // If token refresh fails, clear token and reject
                apiToken.token = null;
                return Promise.reject(refreshError);
            }
        }

        return Promise.reject(error);
    },
);

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
            headers: getAuthHeaders(),
            data: request,
        });
        return {
            message: response.data.message as ApiMessage,
            chatId: response.data.chatId,
            userId: response.data.userId,
            title: response.data.title,
        };
    } catch (error) {
        throw getErrorMessage(error);
    }
}

/** Validates the prompt by sending it to the /validationRes endpoint
 * @param body: object containing userId, chatId, and prompt
 */
export async function validatePrompt(
    request: ApiChatRequest,
): Promise<ApiChatResponse> {
    try {
        const response = await axios({
            method: "post",
            url: "/validate",
            headers: getAuthHeaders(),
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
            headers: getAuthHeaders(),
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
            headers: getAuthHeaders(),
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
            headers: getAuthHeaders(),
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
        url: "/run",
        baseURL,
        headers: getAuthHeaders(),
        data: request,
    });
}

export async function getChatById(
    chatId?: string,
): Promise<ApiGetChatByIdResponse> {
    try {
        const id = chatId ?? chat.id;
        if (!id) {
            throw new Error("chatId is required");
        }

        const response = await axios({
            method: "get",
            url: `/chats/${id}`,
            baseURL: baseURL,
            headers: getAuthHeaders(),
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
            headers: getAuthHeaders(),
            params: {
                testcaseId,
                chatId: chat.id,
                userId: user.id,
            },
        });
        return response.data;
    } catch (err) {
        throw getErrorMessage(err);
    }
}

export async function getApiToken(): Promise<ApiToken> {
    try {
        const response = await axios({
            method: "post",
            url: "auth/generate",
            baseURL: baseURL,
            data: { userId: user.id },
        });
        return response.data as ApiToken;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function updateChatTitle(
    chatId: string,
    title: string,
): Promise<ApiChatSummary> {
    try {
        const response = await axios({
            method: "patch",
            url: `/chats/${chatId}/title`,
            baseURL,
            data: { title },
            headers: getAuthHeaders(),
        });

        return response.data;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function getMedia(testId: string): Promise<ApiMediaResponse> {
    try {
        const response = await axios({
            method: "get",
            url: `test/${testId}/media`,
            baseURL: baseURL,
            headers: getAuthHeaders(),
        });

        return response.data;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

/** Returns presigned S3 URL for video. Uses JSON when backend sends it; otherwise uses redirect Location (no redirect follow to avoid CORS). */
export async function getVideoUrl(testId: string): Promise<string> {
    const response = await axios({
        method: "get",
        url: `test/${testId}/video`,
        baseURL: baseURL,
        headers: { ...getAuthHeaders(), Accept: "application/json" },
        maxRedirects: 0,
        validateStatus: (status) => status === 200 || status === 307,
    });
    if (response.status === 307) {
        const location = response.headers.location;
        if (typeof location === "string") return location;
        throw new Error("Missing Location header in redirect");
    }
    return (response.data as { url: string }).url;
}

/** Returns presigned S3 URL for screenshot. Uses JSON when backend sends it; otherwise uses redirect Location (no redirect follow to avoid CORS). */
export async function getScreenshotUrl(testId: string): Promise<string> {
    const response = await axios({
        method: "get",
        url: `test/${testId}/screenshot`,
        baseURL: baseURL,
        headers: { ...getAuthHeaders(), Accept: "application/json" },
        maxRedirects: 0,
        validateStatus: (status) => status === 200 || status === 307,
    });
    if (response.status === 307) {
        const location = response.headers.location;
        if (typeof location === "string") return location;
        throw new Error("Missing Location header in redirect");
    }
    return (response.data as { url: string }).url;
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
        const response = await axios({
            method: "get",
            url: "/groups",
            baseURL: baseURL,
            headers: getAuthHeaders(),
        });
        return response.data as ApiGroup[];
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function createGroup(
    request: ApiCreateGroupRequest,
): Promise<ApiCreateGroupResponse> {
    try {
        const response = await axios({
            method: "post",
            url: "/groups",
            baseURL: baseURL,
            headers: getAuthHeaders(),
            data: request,
        });
        return response.data as ApiCreateGroupResponse;
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function assignChatToGroups(
    chatId: string,
    groupIds: string[],
): Promise<void> {
    try {
        await axios({
            method: "post",
            url: `/chats/${chatId}/groups`,
            baseURL: baseURL,
            headers: getAuthHeaders(),
            data: { groupIds },
        });
    } catch (error) {
        throw getErrorMessage(error);
    }
}

export async function removeChatFromGroup(
    chatId: string,
    groupId: string,
): Promise<void> {
    try {
        await axios({
            method: "delete",
            url: `/chats/${chatId}/groups/${groupId}`,
            baseURL: baseURL,
            headers: getAuthHeaders(),
        });
    } catch (error) {
        throw getErrorMessage(error);
    }
}
