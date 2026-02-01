import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { Runner } from "../../src/lib/runner.svelte";
import * as shared from "../../src/lib/shared.svelte";
import * as api from "../../src/lib/api";
import { toast } from "svelte-sonner";

// Mock dependencies
vi.mock("svelte-sonner");
vi.mock("../../src/lib/shared.svelte", () => ({
    user: { id: "" },
    chat: { id: "", isLoading: false },
    messages: [],
    ChatDate: { Range: undefined },
    baseURL: "http://localhost:8081/api/v1",
}));
vi.mock("../../src/lib/api");

describe("Runner", () => {
    let runner: Runner;
    const mockUserId = "auth0|user123";
    const mockChatId = "chat456";
    const mockTestId = "test789";

    beforeEach(() => {
        vi.clearAllMocks();
        vi.useFakeTimers();

        // Reset shared state
        shared.user.id = mockUserId;
        shared.chat.id = mockChatId;

        runner = new Runner(mockChatId, mockUserId);
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    describe("isRunning", () => {
        it("should return false initially", () => {
            expect(runner.isRunning()).toBe(false);
        });

        it.skip("should return true when a test is running", async () => {
            // Setup for run to succeed
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;
            await runner.setTest(mockTestId);

            // Start run in background (don't await)
            const runPromise = runner.run();

            // Advance timers slightly to enter running state
            await vi.advanceTimersByTimeAsync(100);

            expect(runner.isRunning()).toBe(true);

            // Complete the run
            await vi.advanceTimersByTimeAsync(5000);
            await runPromise;

            expect(runner.isRunning()).toBe(false);
        });
    });

    describe("setTest", () => {
        it("should set the test ID when not running", async () => {
            await runner.setTest(mockTestId);
            expect(runner.getCurTest()).toBe(mockTestId);
        });

        it.skip("should throw error when trying to set test while running", async () => {
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;
            await runner.setTest(mockTestId);

            // Start running
            const runPromise = runner.run();
            await vi.advanceTimersByTimeAsync(100);

            // Try to set test while running
            await expect(runner.setTest("another-test")).rejects.toThrow(
                "Es läuft momentan ein Test",
            );

            // Cleanup
            await vi.advanceTimersByTimeAsync(5000);
            await runPromise;
        });

        it("should allow setting test after previous test completes", async () => {
            await runner.setTest(mockTestId);
            expect(runner.getCurTest()).toBe(mockTestId);

            await runner.setTest("new-test-id");
            expect(runner.getCurTest()).toBe("new-test-id");
        });
    });

    describe("storeTest", () => {
        const mockTestCode = "const test = 'example';";
        const mockSaveResponse = {
            testcaseId: mockTestId,
            action: "saved",
        };

        beforeEach(() => {
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;
        });

        it.skip("should save test successfully with sanitized user ID", async () => {
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue(mockSaveResponse);

            const storePromise = runner.storeTest(mockTestCode);

            // Verify state is "saving"
            expect(runner.getStorageState()).toBe("saving");

            // Advance through sleep timer
            await vi.advanceTimersByTimeAsync(3000);
            await storePromise;

            // Verify API was called with sanitized user ID
            expect(saveTestLocalMock).toHaveBeenCalledWith({
                userId: "user123", // sanitized from "auth0|user123"
                chatId: mockChatId,
                code: mockTestCode,
            });

            // Verify success toast
            expect(toast.success).toHaveBeenCalledWith(
                "Test erfolgreich gespeichert!",
            );

            // Verify state is "success"
            expect(runner.getStorageState()).toBe("success");
            expect(runner.getCurTest()).toBe(mockTestId);

            // Advance to reset state
            await vi.advanceTimersByTimeAsync(2000);
            expect(runner.getStorageState()).toBe("idle");
        });

        it.skip("should save test with unsanitized user ID when no pipe present", async () => {
            shared.user.id = "simpleUserId";
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue(mockSaveResponse);

            const storePromise = runner.storeTest(mockTestCode);
            await vi.advanceTimersByTimeAsync(3000);
            await storePromise;

            expect(saveTestLocalMock).toHaveBeenCalledWith({
                userId: "simpleUserId", // no sanitization needed
                chatId: mockChatId,
                code: mockTestCode,
            });
        });

        it("should handle missing user ID", async () => {
            const runnerWithNoUser = new Runner(mockChatId, "");

            await runnerWithNoUser.storeTest(mockTestCode);

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: "Benutzer- oder Konversations-ID fehlt.",
                },
            );
            expect(runnerWithNoUser.getStorageState()).toBe("idle");
        });

        it("should handle missing chat ID", async () => {
            const runnerWithNoChat = new Runner("", mockUserId);

            await runnerWithNoChat.storeTest(mockTestCode);

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: "Benutzer- oder Konversations-ID fehlt.",
                },
            );
            expect(runnerWithNoChat.getStorageState()).toBe("idle");
        });

        it.skip("should handle API error with error instance", async () => {
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            const errorMessage = "Network error occurred";
            saveTestLocalMock.mockRejectedValue(new Error(errorMessage));

            const storePromise = runner.storeTest(mockTestCode);

            expect(runner.getStorageState()).toBe("saving");

            // Wait for promise to complete first
            await storePromise;

            expect(runner.getStorageState()).toBe("error");
            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: errorMessage,
                },
            );

            // Verify state resets to idle after timeout
            await vi.advanceTimersByTimeAsync(2000);
            expect(runner.getStorageState()).toBe("idle");
        });

        it.skip("should handle API error with non-Error object", async () => {
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockRejectedValue("String error");

            const storePromise = runner.storeTest(mockTestCode);
            await storePromise;

            expect(runner.getStorageState()).toBe("error");
            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: "Unbekannter Fehler",
                },
            );

            await vi.advanceTimersByTimeAsync(2000);
            expect(runner.getStorageState()).toBe("idle");
        });

        it.skip("should reset storage state to idle after success timeout", async () => {
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue(mockSaveResponse);

            const storePromise = runner.storeTest(mockTestCode);
            await vi.advanceTimersByTimeAsync(3000);
            await storePromise;

            expect(runner.getStorageState()).toBe("success");

            // Advance to trigger finally timeout
            await vi.advanceTimersByTimeAsync(2000);
            expect(runner.getStorageState()).toBe("idle");
        });
    });

    describe("getCurTest", () => {
        it("should return empty string initially", () => {
            expect(runner.getCurTest()).toBe("");
        });

        it("should return the stored test ID", async () => {
            await runner.setTest(mockTestId);
            expect(runner.getCurTest()).toBe(mockTestId);
        });
    });

    describe("getStorageState", () => {
        it("should return 'idle' initially", () => {
            expect(runner.getStorageState()).toBe("idle");
        });

        it.skip("should return current storage state", async () => {
            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue({
                testcaseId: mockTestId,
                action: "saved",
            });

            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;

            const storePromise = runner.storeTest("test code");

            expect(runner.getStorageState()).toBe("saving");

            await vi.advanceTimersByTimeAsync(3000);
            await storePromise;

            expect(runner.getStorageState()).toBe("success");
        });
    });

    describe("run", () => {
        beforeEach(async () => {
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;
            await runner.setTest(mockTestId);
        });

        it.skip("should execute test successfully", async () => {
            const runPromise = runner.run();

            // Advance timers to allow mutex and toast to execute
            await vi.advanceTimersByTimeAsync(100);

            // Verify start message
            expect(toast.message).toHaveBeenCalledWith("Test wird ausgeführt", {
                class: "!bg-purple",
                style: "!bg-red",
                description: `Id: ${mockTestId}`,
            });

            expect(runner.isRunning()).toBe(true);

            // Advance through execution
            await vi.advanceTimersByTimeAsync(5000);
            await runPromise;

            // Verify completion
            expect(runner.result).toBe("asaskjjsdkakjdashdjaskjdhaskjdhjaskd");
            expect(runner.isRunning()).toBe(false);
            expect(toast.message).toHaveBeenCalledWith(
                "Testausführung beendet",
            );
        });

        it("should handle missing user ID", async () => {
            const runnerWithNoUser = new Runner(mockChatId, "");

            await runnerWithNoUser.run();

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description:
                        "Benutzer-, -Konversations oder Test-ID fehlt.",
                },
            );
            expect(runnerWithNoUser.isRunning()).toBe(false);
        });

        it("should handle missing chat ID", async () => {
            const runnerWithNoChat = new Runner("", mockUserId);

            await runnerWithNoChat.run();

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description:
                        "Benutzer-, -Konversations oder Test-ID fehlt.",
                },
            );
            expect(runnerWithNoChat.isRunning()).toBe(false);
        });

        it("should handle missing test ID", async () => {
            runner = new Runner(mockChatId, mockUserId); // Fresh runner without test set

            await runner.run();

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description:
                        "Benutzer-, -Konversations oder Test-ID fehlt.",
                },
            );
            expect(runner.isRunning()).toBe(false);
        });

        it.skip("should prevent concurrent test execution", async () => {
            const firstRun = runner.run();
            await vi.advanceTimersByTimeAsync(100);

            expect(runner.isRunning()).toBe(true);

            // Try to run again
            await runner.run();

            expect(toast.error).toHaveBeenCalledWith(
                "Es läuft bereits ein Test",
                {
                    description: `Id: ${mockTestId}`,
                },
            );

            // Complete first run
            await vi.advanceTimersByTimeAsync(5000);
            await firstRun;
        });

        it.skip("should allow running test again after completion", async () => {
            // First run
            const firstRun = runner.run();
            await vi.advanceTimersByTimeAsync(5100);
            await firstRun;

            expect(runner.isRunning()).toBe(false);

            // Clear previous mocks
            vi.clearAllMocks();

            // Second run should work
            const secondRun = runner.run();

            // Advance to allow toast to be called
            await vi.advanceTimersByTimeAsync(100);

            expect(toast.message).toHaveBeenCalledWith("Test wird ausgeführt", {
                class: "!bg-purple",
                style: "!bg-red",
                description: `Id: ${mockTestId}`,
            });

            await vi.advanceTimersByTimeAsync(5000);
            await secondRun;

            expect(runner.isRunning()).toBe(false);
        });

        it.skip("should update result property after execution", async () => {
            expect(runner.result).toBe("");

            const runPromise = runner.run();
            await vi.advanceTimersByTimeAsync(5000);
            await runPromise;

            expect(runner.result).toBe("asaskjjsdkakjdashdjaskjdhaskjdhjaskd");
        });

        it("should handle all missing IDs in error message", async () => {
            shared.user.id = "";
            shared.chat.id = "";
            runner = new Runner("", "");

            const consoleSpy = vi
                .spyOn(console, "error")
                .mockImplementation(() => {});

            await runner.run();

            expect(consoleSpy).toHaveBeenCalledWith(
                expect.stringContaining("Missing IDs - ChatID:"),
            );
            expect(consoleSpy).toHaveBeenCalledWith(
                expect.stringContaining("UserID:"),
            );
            expect(consoleSpy).toHaveBeenCalledWith(
                expect.stringContaining("TestID:"),
            );

            consoleSpy.mockRestore();
        });
    });

    describe("Mutex behavior", () => {
        it("should properly synchronize setTest calls", async () => {
            const promise1 = runner.setTest("test1");
            const promise2 = runner.setTest("test2");
            const promise3 = runner.setTest("test3");

            await Promise.all([promise1, promise2, promise3]);

            // The last call should win
            expect(runner.getCurTest()).toBe("test3");
        });
    });

    describe("Edge cases", () => {
        it.skip("should handle rapid state changes", async () => {
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;

            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue({
                testcaseId: "test1",
                action: "saved",
            });

            // Start multiple store operations
            const store1 = runner.storeTest("code1");

            expect(runner.getStorageState()).toBe("saving");

            await vi.advanceTimersByTimeAsync(3000);
            await store1;

            expect(runner.getStorageState()).toBe("success");
        });

        it.skip("should preserve result across multiple runs", async () => {
            shared.user.id = mockUserId;
            shared.chat.id = mockChatId;
            await runner.setTest(mockTestId);

            // First run
            const run1 = runner.run();
            await vi.advanceTimersByTimeAsync(5000);
            await run1;

            const firstResult = runner.result;
            expect(firstResult).toBe("asaskjjsdkakjdashdjaskjdhaskjdhjaskd");

            // Second run
            const run2 = runner.run();
            await vi.advanceTimersByTimeAsync(5000);
            await run2;

            // Result should be updated (same in this case)
            expect(runner.result).toBe("asaskjjsdkakjdashdjaskjdhaskjdhjaskd");
        });

        it.skip("should handle user ID with multiple pipe characters", async () => {
            shared.user.id = "provider|sub|extra";
            shared.chat.id = mockChatId;

            const saveTestLocalMock = vi.mocked(api.saveTestLocal);
            saveTestLocalMock.mockResolvedValue({
                testcaseId: mockTestId,
                action: "saved",
            });

            const storePromise = runner.storeTest("test");
            await vi.advanceTimersByTimeAsync(3000);
            await storePromise;

            // split() takes only the second part after first pipe
            expect(saveTestLocalMock).toHaveBeenCalledWith({
                userId: "sub",
                chatId: mockChatId,
                code: "test",
            });
        });
    });

    describe("Initial state", () => {
        it("should have correct initial values", () => {
            const newRunner = new Runner(mockChatId, mockUserId);

            expect(newRunner.isRunning()).toBe(false);
            expect(newRunner.getCurTest()).toBe("");
            expect(newRunner.getStorageState()).toBe("idle");
            expect(newRunner.result).toStrictEqual([]);
        });
    });
});
