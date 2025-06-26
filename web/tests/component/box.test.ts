import { render, screen } from "@testing-library/svelte";
import { expect, test } from "vitest";
import Box from "../../src/lib/Box.svelte";

test("Box shows message from Bot", async () => {
    render(Box, { msg: "This is a Message from User!", name: "User" });

    const name = screen.getByText("User");
    const content = screen.getByText("This is a Message from User!");

    expect(name).toBeInTheDocument();
    expect(content).toBeInTheDocument();

    const parentDiv = name.parentElement;
    const grandparentDiv = parentDiv.parentElement;

    // Classes Check
    expect(content).toHaveClass("text-end", "font-sans", "whitespace-pre-wrap");

    expect(name).toHaveClass(
        "text-end",
        "tracking-wide",
        "uppercase",
        "font-bold",
        "text-xl",
    );
    expect(parentDiv).toHaveClass(
        "font-mono",
        "bg-sky-300",
        "w-fit",
        "justify-end",
        "p-2.5",
        "border-2",
        " border-black ",
        "border-solid ",
        "rounded-xl",
    );
    expect(grandparentDiv).toHaveClass("flex", "justify-end", "m-4");
});

test("Box shows message from Bot", async () => {
    render(Box, { msg: "This is a message from Bot!", name: "Bot" });
    const name = screen.getByText("Bot");
    const content = screen.getByText("This is a message from Bot!");
    expect(name).toBeInTheDocument();
    expect(content).toBeInTheDocument();

    const parentDiv = name.parentElement;
    const grandparentDiv = parentDiv.parentElement;
    // Classes Check for Bot
    expect(content).toHaveClass(
        "text-start",
        "font-sans",
        "whitespace-pre-wrap",
    );

    expect(name).toHaveClass(
        "text-start",
        "tracking-wide",
        "uppercase",
        "font-bold",
        "text-xl",
    );

    expect(parentDiv).toHaveClass(
        "font-mono",
        "bg-gray-200",
        "w-fit",
        "justify-start",
        "p-2.5",
        "border-2",
        "border-black",
        "border-solid",
        "rounded-xl",
    );

    expect(grandparentDiv).toHaveClass("flex", "justify-start", "m-4");
});

test("Box applies correct dynamic classes", async () => {
    const container = render(Box, {
        msg: "Test Message",
        name: "Test",
    });

    const name = screen.getByText("Test");
    const content = screen.getByText("Test Message");

    expect(name).toBeInTheDocument();
    expect(content).toBeInTheDocument();

    const parentDiv = name.parentElement;
    const grandparentDiv = parentDiv.parentElement;
    // Classes Check for Bot
    expect(content).toHaveClass("font-sans", "whitespace-pre-wrap");

    expect(name).toHaveClass(
        "tracking-wide",
        "uppercase",
        "font-bold",
        "text-xl",
    );

    expect(parentDiv).toHaveClass(
        "font-mono",
        "w-fit",
        "p-2.5",
        "border-2",
        "border-black",
        "border-solid",
        "rounded-xl",
    );

    expect(grandparentDiv).toHaveClass("flex", "m-4");
    expect(content).not.toHaveClass("text-start", "text-end");
    expect(parentDiv).not.toHaveClass("justify-start", "justify-end");
    expect(parentDiv).not.toHaveClass("bg-sky-300", "bg-gray-200");
});
