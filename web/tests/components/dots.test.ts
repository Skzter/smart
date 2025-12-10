import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import '@testing-library/jest-dom/vitest';

import Dots from "../../src/lib/components/Dots.svelte";

describe("Dots", () => {
    it("renders the bot icon", () => {
        const { container } = render(Dots);

        const botIcon = container.querySelector('svg');
        expect(botIcon).toBeInTheDocument();
    });

    it("renders three animated dots", () => {
        const { container } = render(Dots);

        const dots = container.querySelectorAll('.animate-bounce');
        expect(dots).toHaveLength(3);
    });

    it("applies different animation delays to each dot", () => {
        const { container } = render(Dots);

        const dots = container.querySelectorAll('.animate-bounce');
        
        expect(dots[0]).toHaveStyle({ 'animation-delay': '0ms' });
        expect(dots[1]).toHaveStyle({ 'animation-delay': '150ms' });
        expect(dots[2]).toHaveStyle({ 'animation-delay': '300ms' });
    });

    it("renders dot characters correctly", () => {
        const { container } = render(Dots);

        const dots = container.querySelectorAll('.animate-bounce');
        
        dots.forEach(dot => {
            expect(dot.textContent).toBe('●');
        });
    });

    it("has correct container layout", () => {
        const { container } = render(Dots);

        const mainContainer = container.querySelector('.flex.justify-start.gap-2');
        expect(mainContainer).toBeInTheDocument();
        
        const dotsContainer = container.querySelector('.flex.gap-1');
        expect(dotsContainer).toBeInTheDocument();
    });

    it("renders bot icon container with correct styling", () => {
        const { container } = render(Dots);

        const iconContainer = container.querySelector('.h-8.w-8.rounded-full.bg-muted');
        expect(iconContainer).toBeInTheDocument();
    });

    it("renders message bubble with correct styling", () => {
        const { container } = render(Dots);

        const messageBubble = container.querySelector('.bg-muted.rounded-2xl.rounded-bl-sm');
        expect(messageBubble).toBeInTheDocument();
    });
});
