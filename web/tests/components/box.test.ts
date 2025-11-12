import { render, screen } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import Box from "../../src/components/Box.svelte";

describe('Box component', () => {
    describe('when the message is from "User"', () => {
        it('renders the message and applies User-specific styles', () => {
          const { container } = render(Box, { 
            msg: 'This is a message from User!', 
            name: 'User',
            userId: 'test-user-123',
            conversationId: 'test-conv-456'
        });

          const nameElement = screen.getByText('User');
          const messageElement = screen.getByText('This is a message from User!');

          expect(nameElement).toBeInTheDocument();
          expect(messageElement).toBeInTheDocument();

          const outerDiv = container.querySelector('.flex.m-4');
          expect(outerDiv).toHaveClass('justify-end');

          const innerDiv = nameElement.closest('.font-mono');
          expect(innerDiv).toHaveClass('bg-sky-300', 'w-[75%]');
    });
    });

    describe('when the message is from "Bot"', () => {
        it('renders the message and applies Bot-specific styles', () => {
            const { container } = render(Box, { 
                msg: 'This is a message from Bot!', 
                name: 'Bot',
                userId: 'test-user-123',
                conversationId: 'test-conv-456'
            });

            const nameElement = screen.getByText("Bot");
            const messageElement = screen.getByText(
                "This is a message from Bot!",
            );

            expect(nameElement).toBeInTheDocument();
            expect(messageElement).toBeInTheDocument();

            const outerDiv = container.querySelector('.flex.m-4');
            expect(outerDiv).toHaveClass("justify-start");

            const innerDiv = nameElement.closest('.font-mono');
            expect(innerDiv).toHaveClass('bg-gray-200', 'w-fit');
        });
    });

    describe('when the sender is not "User" or "Bot"', () => {
        it("renders with default styles", () => {
            const { container } = render(Box, { 
                msg: "Test Message", 
                name: "Test",
                userId: 'test-user-123',
                conversationId: 'test-conv-456'
            });

            const nameElement = screen.getByText("Test");
            const messageElement = screen.getByText("Test Message");

            expect(nameElement).toBeInTheDocument();
            expect(messageElement).toBeInTheDocument();

            const outerDiv = container.querySelector('.flex.m-4');
            expect(outerDiv).not.toHaveClass("justify-end", "justify-start");

            const innerDiv = nameElement.closest('.font-mono');
            expect(innerDiv).not.toHaveClass("bg-sky-300", "bg-gray-200", "w-[75%]", "w-fit");
        });
    });

    describe("when the name is empty", () => {
        it("renders the message with default styles and an empty heading", () => {
            const { container } = render(Box, { 
                msg: "Empty name test", 
                name: "",
                userId: 'test-user-123',
                conversationId: 'test-conv-456'
            });

            const messageElement = screen.getByText("Empty name test");
            expect(messageElement).toBeInTheDocument();

            const nameElement = screen.getByRole("heading", { level: 1 });
            expect(nameElement).toBeInTheDocument();
            expect(nameElement.textContent).toBe("");

            const outerDiv = container.querySelector('.flex.m-4');
            expect(outerDiv).not.toHaveClass("justify-end", "justify-start");

            const innerDiv = nameElement.closest('.font-mono');
            expect(innerDiv).not.toHaveClass("bg-sky-300", "bg-gray-200");
        });
    });

    describe("when the message is empty", () => {
        it("renders the name with an empty message paragraph", () => {
            const { container } = render(Box, { 
                msg: "", 
                name: "User",
                userId: 'test-user-123',
                conversationId: 'test-conv-456'
            });
            const nameElement = screen.getByText("User");
            expect(nameElement).toBeInTheDocument();

            const messageElement = container.querySelector("p.font-sans");
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toBe("");
        });
    });

    describe("when showSave is true", () => {
        it("renders the Save button", () => {
            render(Box, {
                msg: "import { test } from 'playwright';",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
                showSave: true
            });

            const saveButton = screen.getByRole("button", { name: /save/i });
            expect(saveButton).toBeInTheDocument();
        });
    });

    describe("when showSave is false", () => {
        it("does not render the Save button", () => {
            render(Box, {
                msg: "import { test } from 'playwright';",
                name: "Bot",
                userId: "test-user-123",
                conversationId: "test-conv-456",
                showSave: false
            });

            const saveButton = screen.queryByRole("button", { name: /save/i });
            expect(saveButton).not.toBeInTheDocument();
        });
    });
});
