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
