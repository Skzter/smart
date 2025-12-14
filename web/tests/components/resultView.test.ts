import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";

import ResultView from "../../src/lib/components/ResultView.svelte";

describe("ResultView", () => {
    it("renders the main container", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(
            ".w-full.h-full.flex.items-center.justify-center.bg-gray-50",
        );
        expect(mainDiv).toBeInTheDocument();
    });

    it("renders with full width and height", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(".w-full.h-full");
        expect(mainDiv).toBeInTheDocument();
    });

    it("centers content with flex layout", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(
            ".flex.items-center.justify-center",
        );
        expect(mainDiv).toBeInTheDocument();
    });

    it("has gray background color", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(".bg-gray-50");
        expect(mainDiv).toBeInTheDocument();
    });

    it("renders text-center container", () => {
        const { container } = render(ResultView);

        const textCenter = container.querySelector(".text-center");
        expect(textCenter).toBeInTheDocument();
    });

    it("displays placeholder text", () => {
        const { container } = render(ResultView);

        const paragraph = container.querySelector("p");
        expect(paragraph).toBeInTheDocument();
        expect(paragraph?.textContent).toBe("Placeholder für Ergebnisse");
    });

    it("renders paragraph with correct styling", () => {
        const { container } = render(ResultView);

        const paragraph = container.querySelector("p.text-gray-400.text-lg");
        expect(paragraph).toBeInTheDocument();
    });

    it("renders paragraph with gray text color", () => {
        const { container } = render(ResultView);

        const paragraph = container.querySelector("p.text-gray-400");
        expect(paragraph).toBeInTheDocument();
    });

    it("renders paragraph with large text size", () => {
        const { container } = render(ResultView);

        const paragraph = container.querySelector("p.text-lg");
        expect(paragraph).toBeInTheDocument();
    });

    it("has correct nested structure", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(".w-full.h-full");
        const textCenter = mainDiv?.querySelector(".text-center");
        const paragraph = textCenter?.querySelector("p");

        expect(mainDiv).toBeInTheDocument();
        expect(textCenter).toBeInTheDocument();
        expect(paragraph).toBeInTheDocument();
    });

    it("renders exactly one paragraph element", () => {
        const { container } = render(ResultView);

        const paragraphs = container.querySelectorAll("p");
        expect(paragraphs).toHaveLength(1);
    });

    it("renders exactly one div with text-center class", () => {
        const { container } = render(ResultView);

        const textCenterDivs = container.querySelectorAll(".text-center");
        expect(textCenterDivs).toHaveLength(1);
    });

    it("contains all required CSS classes on main container", () => {
        const { container } = render(ResultView);

        const mainDiv = container.firstElementChild;
        expect(mainDiv).toHaveClass("w-full");
        expect(mainDiv).toHaveClass("h-full");
        expect(mainDiv).toHaveClass("flex");
        expect(mainDiv).toHaveClass("items-center");
        expect(mainDiv).toHaveClass("justify-center");
        expect(mainDiv).toHaveClass("bg-gray-50");
    });

    it("contains all required CSS classes on paragraph", () => {
        const { container } = render(ResultView);

        const paragraph = container.querySelector("p");
        expect(paragraph).toHaveClass("text-gray-400");
        expect(paragraph).toHaveClass("text-lg");
    });

    it("renders without any props", () => {
        const { container } = render(ResultView);

        expect(container.firstElementChild).toBeInTheDocument();
    });

    it("displays German text for placeholder", () => {
        const { container } = render(ResultView);

        expect(container.textContent).toContain("Placeholder für Ergebnisse");
    });

    it("renders component structure correctly", () => {
        const { container } = render(ResultView);

        const mainDiv = container.querySelector(
            "div.w-full.h-full",
        ) as HTMLElement;
        const innerDiv = container.querySelector(
            "div.text-center",
        ) as HTMLElement;
        const text = container.querySelector("p.text-gray-400") as HTMLElement;

        expect(mainDiv).toContainElement(innerDiv);
        expect(innerDiv).toContainElement(text);
    });
});
