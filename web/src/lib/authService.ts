import { writable, get } from "svelte/store";
import { UserManager, WebStorageStateStore, type User, type UserProfile } from "oidc-client-ts";

interface AuthState {
    isAuthenticated: boolean;
    user: UserProfile | null;
    userManager: UserManager | null;
}

const createAuthStore = () => {
    const { subscribe, set, update } = writable<AuthState>({
        isAuthenticated: false,
        user: null,
        userManager: null,
    });

    const initAuth = async () => {
        const response = await fetch("/auth_config.json");
        const authConfig = await response.json();

        const userManager = new UserManager({
            authority: authConfig.authority,
            client_id: authConfig.client_id,
            redirect_uri: authConfig.redirect_uri || window.location.origin,
            response_type: authConfig.response_type || "code",
            scope: authConfig.scope || "openid profile email",
            userStore: new WebStorageStateStore({ store: window.localStorage }),
            loadUserInfo: true,
        });

        // Handle callback if code is present
        if (
            window.location.search.includes("code=") &&
            window.location.search.includes("state=")
        ) {
            try {
                await userManager.signinCallback();
                // Clear query params
                window.history.replaceState({}, document.title, window.location.pathname);
            } catch (err) {
                console.error("Error handling redirect callback:", err);
            }
        }

        let user: User | null = null;
        try {
            user = await userManager.getUser();
        } catch (err) {
            console.error("Error getting user:", err);
        }

        set({
            isAuthenticated: !!user && !user.expired,
            user: user?.profile || null,
            userManager,
        });

        // Setup events to update store
        userManager.events.addUserLoaded((loadedUser) => {
            update(state => ({
                ...state,
                isAuthenticated: true,
                user: loadedUser.profile
            }));
        });

        userManager.events.addUserUnloaded(() => {
            update(state => ({
                ...state,
                isAuthenticated: false,
                user: null
            }));
        });
        
        userManager.events.addAccessTokenExpired(() => {
             update(state => ({
                ...state,
                isAuthenticated: false,
                user: null
            }));
        });
    };

    const login = async () => {
        const manager = get({ subscribe }).userManager;
        if (manager) {
            await manager.signinRedirect();
        }
    };

    const logout = async () => {
        const manager = get({ subscribe }).userManager;
        if (manager) {
            await manager.signoutRedirect();
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
