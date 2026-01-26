/// <reference types="vitest/globals" />
import * as matchers from "@testing-library/jest-dom/matchers";
import { expect } from "vitest";

expect.extend(matchers);

Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: vi.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(), // deprecated
        removeListener: vi.fn(), // deprecated
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
    })),
});

if (!Element.prototype.animate) {
    Element.prototype.animate = () => ({
        cancel: () => {},
        finish: () => {},
        play: () => {},
        pause: () => {},
        reverse: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        finished: Promise.resolve(),
        playState: "finished",
    });
}

// Mock HTMLDialogElement methods for jsdom
HTMLDialogElement.prototype.showModal = vi.fn(function (
    this: HTMLDialogElement,
) {
    this.open = true;
});
HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
    this.open = false;
});

// Mock document.queryCommandSupported for Monaco Editor
if (!document.queryCommandSupported) {
    document.queryCommandSupported = vi.fn(() => false);
}

// Mock localStorage for ApiTokenStore
const localStorageMock: Storage = {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
    length: 0,
    key: vi.fn(() => null),
};

Object.defineProperty(global, "localStorage", {
    value: localStorageMock,
    writable: true,
});
// Prevent "document is not defined" errors from bits-ui cleanup timeouts
// by ensuring cleanup waits for pending timers
afterEach(async () => {
    // Wait a bit for any pending timers/cleanup to complete
    await new Promise((resolve) => setTimeout(resolve, 100));
});
