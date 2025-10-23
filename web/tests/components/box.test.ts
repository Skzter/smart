import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import Box from '../../src/components/Box.svelte';

describe('Box component', () => {
	describe('when the message is from "User"', () => {
    	it('renders the message and applies User-specific styles', () => {
      	render(Box, { msg: 'This is a message from User!', name: 'User' });

      	const nameElement = screen.getByText('User');
      	const messageElement = screen.getByText('This is a message from User!');

      	expect(nameElement).toBeInTheDocument();
      	expect(messageElement).toBeInTheDocument();

      	const outerDiv = nameElement.parentElement?.parentElement;
      	expect(outerDiv).toHaveClass('justify-end');

      	const innerDiv = nameElement.parentElement;
      	expect(innerDiv).toHaveClass('bg-sky-300', 'w-[75%]');
    });
	});
	describe('when the message is from "Bot"', () => {
		it('renders the message and applies Bot-specific styles', () => {
			render(Box, { msg: 'This is a message from Bot!', name: 'Bot' });

			const nameElement = screen.getByText('Bot');
			const messageElement = screen.getByText('This is a message from Bot!');

			expect(nameElement).toBeInTheDocument();
			expect(messageElement).toBeInTheDocument();

			const outerDiv = nameElement.parentElement?.parentElement;
			expect(outerDiv).toHaveClass('justify-start');

			const innerDiv = nameElement.parentElement;
      		expect(innerDiv).toHaveClass('bg-gray-200', 'w-fit');
		});
	});

	describe('when the sender is not "User" or "Bot"', () => {
		it('renders with default styles', () => {
			render(Box, { msg: 'Test Message', name: 'Test' });

			const nameElement = screen.getByText('Test');
			const messageElement = screen.getByText('Test Message');

			expect(nameElement).toBeInTheDocument();
			expect(messageElement).toBeInTheDocument();

			const outerDiv = nameElement.parentElement?.parentElement;
			expect(outerDiv).not.toHaveClass('justify-end', 'justify-start');

			const innerDiv = nameElement.parentElement;
			expect(innerDiv).not.toHaveClass('bg-sky-300', 'bg-gray-200');
			expect(innerDiv).not.toHaveClass('justify-end', 'justify-start');

			expect(nameElement).not.toHaveClass('text-end', 'text-start');
			expect(messageElement).not.toHaveClass('text-end', 'text-start');
		});
	});

	describe('when the name is empty', () => {
		it('renders the message with default styles and an empty heading', () => {
			render(Box, { msg: 'Empty name test', name: '' });

			const messageElement = screen.getByText('Empty name test');
			expect(messageElement).toBeInTheDocument();

			const nameElement = screen.getByRole('heading', { level: 1 });
			expect(nameElement).toBeInTheDocument();
			expect(nameElement.textContent).toBe('');

			const outerDiv = nameElement.parentElement?.parentElement;
			expect(outerDiv).not.toHaveClass('justify-end', 'justify-start');

			const innerDiv = nameElement.parentElement;
			expect(innerDiv).not.toHaveClass('bg-sky-300', 'bg-gray-200');

			expect(nameElement).not.toHaveClass('text-end', 'text-start');
			expect(messageElement).not.toHaveClass('text-end', 'text-start');
		});
	});

	describe('when the message is empty', () => {
		it('renders the name with an empty message paragraph', () => {
			const { container } = render(Box, { msg: '', name: 'User' });
			const nameElement = screen.getByText('User');
			expect(nameElement).toBeInTheDocument();

			// The message paragraph should be empty
			const messageElement = container.querySelector('p');
			expect(messageElement).toBeInTheDocument();
			expect(messageElement?.textContent).toBe('');
		});
	});
});
