import { render } from "@testing-library/svelte";
import { describe, it, expect, beforeEach } from "vitest";
import '@testing-library/jest-dom/vitest';

import OutputView from "../../src/lib/components/OutputView.svelte";
import { Runner } from "../../src/lib/runner.svelte";

describe("OutputView", () => {
    let testRunner: Runner;

    beforeEach(() => {
        testRunner = new Runner();
    });

    it("renders the output header", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header).toBeInTheDocument();
    });

    it("displays 'Test Output' text in header", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("Test Output");
    });

    it("displays empty result in header initially", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toMatch(/Test Output\s*$/);
    });

    it("displays result value in header when testRunner has result", () => {
        testRunner.result = "Test completed successfully";
        
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("Test Output Test completed successfully");
    });

    it("renders Terminal.Root component", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const terminal = container.querySelector('.m-0.max-w-none');
        expect(terminal).toBeInTheDocument();
    });

    it("renders Terminal.Root with correct props and classes", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const terminal = container.querySelector('.m-0.max-w-none.h-full.rounded-none.border-none');
        expect(terminal).toBeInTheDocument();
    });

    it("renders complete component structure", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const mainContainer = container.querySelector('.flex.flex-col');
        expect(mainContainer).toBeInTheDocument();
        
        const header = container.querySelector('.px-4.py-2.bg-muted\\/50.text-sm.font-medium');
        expect(header).toBeInTheDocument();
        
        const contentArea = container.querySelector('.flex-1.overflow-auto');
        expect(contentArea).toBeInTheDocument();
        
        const terminal = container.querySelector('.m-0.max-w-none');
        expect(terminal).toBeInTheDocument();
    });

    it("renders with different result values", () => {
        const customRunner = new Runner();
        customRunner.result = "Custom result text";
        
        const { container } = render(OutputView, {
            props: { testRunner: customRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("Custom result text");
    });

    it("renders all child components in correct order", () => {
        const { container } = render(OutputView, {
            props: { testRunner }
        });

        const mainContainer = container.querySelector('.flex.flex-col');
        const children = mainContainer?.children;
        
        expect(children).toHaveLength(2);
        expect(children?.[0]).toHaveClass('px-4');
        expect(children?.[1]).toHaveClass('flex-1');
    });

    it("renders component with new Runner instance", () => {
        const newRunner = new Runner();
        const { container } = render(OutputView, {
            props: { testRunner: newRunner }
        });

        expect(container.querySelector('.flex.flex-col')).toBeInTheDocument();
    });

    it("handles empty string result", () => {
        const emptyRunner = new Runner();
        emptyRunner.result = "";
        
        const { container } = render(OutputView, {
            props: { testRunner: emptyRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toMatch(/Test Output\s*$/);
    });

    it("handles long result strings", () => {
        const longRunner = new Runner();
        longRunner.result = "This is a very long result string that contains a lot of information about the test execution and its outcomes";
        
        const { container } = render(OutputView, {
            props: { testRunner: longRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("This is a very long result string");
    });

    it("handles special characters in result", () => {
        const specialRunner = new Runner();
        specialRunner.result = "<>&\"'";
        
        const { container } = render(OutputView, {
            props: { testRunner: specialRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("<>&\"'");
    });

    it("handles numeric result values", () => {
        const numericRunner = new Runner();
        numericRunner.result = "123456789";
        
        const { container } = render(OutputView, {
            props: { testRunner: numericRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("123456789");
    });

    it("handles multiline result strings", () => {
        const multilineRunner = new Runner();
        multilineRunner.result = "Line 1\nLine 2\nLine 3";
        
        const { container } = render(OutputView, {
            props: { testRunner: multilineRunner }
        });

        const header = container.querySelector('.px-4.py-2.bg-muted\\/50');
        expect(header?.textContent).toContain("Line 1");
    });
});
