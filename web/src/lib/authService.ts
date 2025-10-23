import { writable, get } from "svelte/store";
import {
    createAuth0Client,
    type Auth0Client,
    type User,
} from "@auth0/auth0-spa-js";

interface AuthState {
    isAuthenticated: boolean;
    user: User | null;
    auth0Client: Auth0Client | null;
}

const createAuthStore = () => {
    const { subscribe, set } = writable<AuthState>({
        isAuthenticated: false,
        user: null,
        auth0Client: null,
    });

    const initAuth = async () => {
        const response = await fetch("/auth_config.json");
        const authConfig = await response.json();
        const auth0Client = await createAuth0Client({
            domain: authConfig.domain,
            clientId: authConfig.clientId,
            cacheLocation: "localstorage",
            authorizationParams: {
                redirect_uri: window.location.origin,
            },
        });

        if (
            window.location.search.includes("code=") &&
            window.location.search.includes("state=")
        ) {
            await auth0Client.handleRedirectCallback();
            window.history.replaceState({}, document.title, "/");
        }

        const isAuthenticated = await auth0Client.isAuthenticated();
        const user = (await auth0Client.getUser()) || null;

        set({
            isAuthenticated,
            user,
            auth0Client,
        });
    };

    const login = async () => {
        const auth0Client = get({ subscribe }).auth0Client;
        if (auth0Client) {
            await auth0Client.loginWithRedirect();
        }
    };

    const logout = () => {
        const auth0Client = get({ subscribe }).auth0Client;
        if (auth0Client) {
            auth0Client.logout({
                logoutParams: {
                    returnTo: window.location.origin,
                },
            });
        }
    };

    return {
        subscribe,
        initAuth,
        login,
        logout,
    };
};

export const auth = createAuthStore();
