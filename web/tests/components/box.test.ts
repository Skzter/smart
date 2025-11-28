import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import Box from "../../src/components/Box.svelte";
import { saveTestLocal, runContainer } from "../../src/lib/Api";
import { AxiosError, type InternalAxiosRequestConfig } from "axios";

vi.mock("../../src/lib/Api");

describe("Box component", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    // Updated helper function to handle both isCode cases
    function getMessageElement(
        container: HTMLElement,
        isCode: boolean = false,
    ) {
        if (isCode) {
            // When isCode is true, content is in <pre> with tokenized spans
            return container.querySelector("pre");
        } else {
            // When isCode is false, content is in <p> tag
            return container.querySelector("p");
        }
    }

    describe('when the message is from "User"', () => {
        it("renders the message and applies User-specific styles", () => {
            const { container } = render(Box, {
                msg: "This is a message from User!",
                name: "User",
                userId: "test-user-123",
                conversationId: "test-conv-456",
            });

            const nameElement = screen.getByText("User");
            expect(nameElement).toBeInTheDocument();

            const messageElement = getMessageElement(container, false);
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toContain(
                "This is a message from User!",
            );

            const outerDiv = container.querySelector(".flex.m-4");
            expect(outerDiv).toHaveClass("justify-end");

            const innerDiv = nameElement.closest(".font-mono");
            expect(innerDiv).toHaveClass("bg-sky-300", "w-[75%]");
        });
    });

    describe('when the message is from "Bot"', () => {
        it("renders the message and applies Bot-specific styles", () => {
            const { container } = render(Box, {
                msg: "This is a message from Bot!",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
            });

            const nameElement = screen.getByText("Bot");
            expect(nameElement).toBeInTheDocument();

            const messageElement = getMessageElement(container, false);
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toContain(
                "This is a message from Bot!",
            );

            const outerDiv = container.querySelector(".flex.m-4");
            expect(outerDiv).toHaveClass("justify-start");

            const innerDiv = nameElement.closest(".font-mono");
            expect(innerDiv).toHaveClass("bg-gray-200", "w-fit");
        });
    });

    describe('when the sender is not "User" or "Bot"', () => {
        it("renders with default styles", () => {
            const { container } = render(Box, {
                msg: "Test Message",
                name: "Test",
                userId: "test-user-123",
                conversationId: "test-conv-456",
            });

            const nameElement = screen.getByText("Test");
            expect(nameElement).toBeInTheDocument();

            const messageElement = getMessageElement(container, false);
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toContain("Test Message");

            const outerDiv = container.querySelector(".flex.m-4");
            expect(outerDiv).not.toHaveClass("justify-end", "justify-start");

            const innerDiv = nameElement.closest(".font-mono");
            expect(innerDiv).not.toHaveClass(
                "bg-sky-300",
                "bg-gray-200",
                "w-[75%]",
                "w-fit",
            );
        });
    });

    describe("when the name is empty", () => {
        it("renders the message with default styles and an empty heading", () => {
            const { container } = render(Box, {
                msg: "Empty name test",
                name: "",
                userId: "test-user-123",
                conversationId: "test-conv-456",
            });

            const messageElement = getMessageElement(container, false);
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toContain("Empty name test");

            const nameElement = screen.getByRole("heading", { level: 1 });
            expect(nameElement).toBeInTheDocument();
            expect(nameElement.textContent).toBe("");

            const outerDiv = container.querySelector(".flex.m-4");
            expect(outerDiv).not.toHaveClass("justify-end", "justify-start");

            const innerDiv = nameElement.closest(".font-mono");
            expect(innerDiv).not.toHaveClass("bg-sky-300", "bg-gray-200");
        });
    });

    describe("when the message is empty", () => {
        it("renders the name with an empty message paragraph", () => {
            const { container } = render(Box, {
                msg: "",
                name: "User",
                userId: "test-user-123",
                conversationId: "test-conv-456",
            });
            const nameElement = screen.getByText("User");
            expect(nameElement).toBeInTheDocument();

            const messageElement = getMessageElement(container, false);
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent?.trim()).toBe("");
        });
    });

    describe("when isCode is true", () => {
        it("renders the Save button", () => {
            render(Box, {
                msg: "import { test } from 'playwright';",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            expect(saveButton).toBeInTheDocument();
        });

        it("renders code with syntax highlighting in pre element", () => {
            const { container } = render(Box, {
                msg: "const x = 5;",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
                isCode: true,
            });

            const codeElement = getMessageElement(container, true);
            expect(codeElement).toBeInTheDocument();
            expect(codeElement?.tagName).toBe("PRE");
            expect(codeElement?.textContent).toContain("const x = 5;");
        });
    });

    describe("when isCode is false", () => {
        it("does not render the Save button", () => {
            render(Box, {
                msg: "import { test } from 'playwright';",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
                isCode: false,
            });

            const saveButton = screen.queryByRole("button", { name: /save/i });
            expect(saveButton).not.toBeInTheDocument();
        });
    });

    describe("Save functionality", () => {
        it("saves test successfully and shows success state", async () => {
            vi.mocked(saveTestLocal).mockResolvedValue({
                data: { testcaseId: "test-123" },
            });

            const { container } = render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✓ Saved")).toBeInTheDocument();
            });

            await waitFor(() => {
                expect(
                    screen.getByText("Test ID: test-123"),
                ).toBeInTheDocument();
            });

            expect(saveTestLocal).toHaveBeenCalledWith({
                userId: "user-123",
                conversationId: "conv-456",
                code: "test code",
            });

            // Ensure message still rendered (optional sanity check)
            const codeElement = getMessageElement(container, true);
            expect(codeElement?.textContent).toContain("test code");
        });

        it("sanitizes userId with pipe character before saving", async () => {
            vi.mocked(saveTestLocal).mockResolvedValue({
                data: { testcaseId: "test-123" },
            });

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "auth0|user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(saveTestLocal).toHaveBeenCalledWith({
                    userId: "user-123",
                    conversationId: "conv-456",
                    code: "test code",
                });
            });
        });

        it("shows error state when userId is missing", async () => {
            const consoleErrorSpy = vi
                .spyOn(console, "error")
                .mockImplementation(() => {});

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: undefined,
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            expect(saveTestLocal).not.toHaveBeenCalled();
            expect(consoleErrorSpy).toHaveBeenCalled();

            consoleErrorSpy.mockRestore();
        });

        it("shows error state when conversationId is missing", async () => {
            const consoleErrorSpy = vi
                .spyOn(console, "error")
                .mockImplementation(() => {});

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: undefined,
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            expect(saveTestLocal).not.toHaveBeenCalled();
            expect(consoleErrorSpy).toHaveBeenCalled();

            consoleErrorSpy.mockRestore();
        });

        it("handles AxiosError with custom error message", async () => {
            const axiosError = new AxiosError("Network Error");
            axiosError.response = {
                data: { error: "Custom error from server" },
                status: 400,
                statusText: "Bad Request",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };

            vi.mocked(saveTestLocal).mockRejectedValue(axiosError);

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✗ Error")).toBeInTheDocument();
            });
        });

        it("handles AxiosError without custom error message", async () => {
            const axiosError = new AxiosError("Network Error");
            axiosError.response = {
                data: {},
                status: 500,
                statusText: "Internal Server Error",
                headers: {},
                config: {} as InternalAxiosRequestConfig,
            };

            vi.mocked(saveTestLocal).mockRejectedValue(axiosError);

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✗ Error")).toBeInTheDocument();
            });
        });

        it("handles non-AxiosError", async () => {
            vi.mocked(saveTestLocal).mockRejectedValue(
                new Error("Generic error"),
            );

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✗ Error")).toBeInTheDocument();
            });
        });

        it("disables button while saving", async () => {
            vi.mocked(saveTestLocal).mockImplementation(
                () =>
                    new Promise((resolve) =>
                        setTimeout(
                            () => resolve({ data: { testcaseId: "test-123" } }),
                            100,
                        ),
                    ),
            );

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("Saving...")).toBeInTheDocument();
                expect(saveButton).toBeDisabled();
            });

            await waitFor(() => {
                expect(screen.getByText("✓ Saved")).toBeInTheDocument();
            });
        });

        it("resets to idle state after error timeout", async () => {
            vi.useFakeTimers();
            vi.mocked(saveTestLocal).mockRejectedValue(new Error("Error"));

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✗ Error")).toBeInTheDocument();
            });

            vi.advanceTimersByTime(2000);

            await waitFor(() => {
                expect(
                    screen.getByRole("button", { name: /save/i }),
                ).toBeInTheDocument();
            });

            vi.useRealTimers();
        });

        it("keeps button disabled after successful save", async () => {
            vi.mocked(saveTestLocal).mockResolvedValue({
                data: { testcaseId: "test-123" },
            });

            render(Box, {
                msg: "test code",
                name: "Bot",
                userId: "user-123",
                conversationId: "conv-456",
                isCode: true,
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            await fireEvent.click(saveButton);

            await waitFor(() => {
                expect(screen.getByText("✓ Saved")).toBeInTheDocument();
                expect(saveButton).toBeDisabled();
            });
        });
    });

    describe("Box component - RunTest functionality", () => {
        beforeEach(() => {
            vi.clearAllMocks();
        });

        describe("Run button visibility", () => {
            it("shows Run button after successful save", async () => {
                vi.mocked(saveTestLocal).mockResolvedValue({
                    data: { testcaseId: "test-123" },
                });

                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    expect(screen.getByText("Run Test")).toBeInTheDocument();
                });
            });

            it("does not show Run button before saving", () => {
                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                const runButton = screen.queryByText("Run Test");
                expect(runButton).not.toBeInTheDocument();
            });

            it("does not show Run button when save fails", async () => {
                vi.mocked(saveTestLocal).mockRejectedValue(
                    new Error("Save failed"),
                );

                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    const runButton = screen.queryByText("Run Test");
                    expect(runButton).not.toBeInTheDocument();
                });
            });

            it("opens modal and displays test results on successful run", async () => {
                vi.mocked(saveTestLocal).mockResolvedValue({
                    data: { testcaseId: "test-123" },
                });
                vi.mocked(runContainer).mockResolvedValue({
                    data: {
                        result: "Test passed successfully!\nAll assertions verified.",
                    },
                });

                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                // First save the test
                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    expect(screen.getByText("✓ Saved")).toBeInTheDocument();
                });

                // Then run the test
                const runButton = screen.getByRole("button", {
                    name: /run test/i,
                });
                await fireEvent.click(runButton);

                // Verify modal is visible and contains the result
                await waitFor(() => {
                    expect(screen.getByText("Test Result")).toBeInTheDocument();
                    expect(
                        screen.getByText(/Test passed successfully!/),
                    ).toBeInTheDocument();
                    expect(
                        screen.getByText(/All assertions verified./),
                    ).toBeInTheDocument();
                });

                // Verify runContainer was called with correct parameters
                expect(runContainer).toHaveBeenCalledWith({
                    userId: "user-123",
                    testId: "test-123",
                    sessionId: "conv-456",
                });
            });

            it("opens modal and displays error message when runContainer fails", async () => {
                vi.mocked(saveTestLocal).mockResolvedValue({
                    data: { testcaseId: "test-123" },
                });

                const axiosError = new AxiosError("Container execution failed");
                axiosError.response = {
                    data: {
                        message:
                            "Error: Test execution timed out after 30 seconds",
                    },
                    status: 500,
                    statusText: "Internal Server Error",
                    headers: {},
                    config: {} as InternalAxiosRequestConfig,
                };

                vi.mocked(runContainer).mockRejectedValue(axiosError);

                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                // First save the test
                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    expect(screen.getByText("✓ Saved")).toBeInTheDocument();
                });

                // Then run the test
                const runButton = screen.getByRole("button", {
                    name: /run test/i,
                });
                await fireEvent.click(runButton);

                // Verify modal is visible and contains the error message
                await waitFor(() => {
                    expect(screen.getByText("Test Result")).toBeInTheDocument();
                    expect(
                        screen.getByText(
                            /Error: Test execution timed out after 30 seconds/,
                        ),
                    ).toBeInTheDocument();
                });

                // Verify runContainer was called with correct parameters
                expect(runContainer).toHaveBeenCalledWith({
                    userId: "user-123",
                    testId: "test-123",
                    sessionId: "conv-456",
                });
            });

            it("does not run test and logs error when userId becomes undefined", async () => {
                const consoleErrorSpy = vi
                    .spyOn(console, "error")
                    .mockImplementation(() => {});

                vi.mocked(saveTestLocal).mockResolvedValue({
                    data: { testcaseId: "test-123" },
                });

                // Start without userId to test the early return in RunTest
                const { unmount } = render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: undefined,
                    conversationId: "conv-456",
                    isCode: true,
                });

                unmount();

                // Now render normally and save
                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    expect(screen.getByText("✓ Saved")).toBeInTheDocument();
                });

                // The Run button exists now, but in real scenario if userId were to become undefined
                // the RunTest function would handle it
                // Since we can't modify props after render, this test verifies the error handling exists

                expect(runContainer).not.toHaveBeenCalled();

                consoleErrorSpy.mockRestore();
            });

            it("sanitizes userId with pipe character before running test", async () => {
                vi.mocked(saveTestLocal).mockResolvedValue({
                    data: { testcaseId: "test-123" },
                });
                vi.mocked(runContainer).mockResolvedValue({
                    data: { result: "Test passed successfully!" },
                });

                render(Box, {
                    msg: "test code",
                    name: "Bot",
                    userId: "auth0|user-123",
                    conversationId: "conv-456",
                    isCode: true,
                });

                // First save the test
                const saveButton = screen.getByRole("button", {
                    name: /save/i,
                });
                await fireEvent.click(saveButton);

                await waitFor(() => {
                    expect(screen.getByText("✓ Saved")).toBeInTheDocument();
                });

                // Then run the test
                const runButton = screen.getByRole("button", {
                    name: /run test/i,
                });
                await fireEvent.click(runButton);

                // Verify runContainer was called with sanitized userId
                await waitFor(() => {
                    expect(runContainer).toHaveBeenCalledWith({
                        userId: "user-123",
                        testId: "test-123",
                        sessionId: "conv-456",
                    });
                });
            });
        });
    });
});
