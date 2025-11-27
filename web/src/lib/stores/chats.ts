import { writable } from "svelte/store";
import type { Writable } from "svelte/store";

export type Chat = {
  id: string;
  title: string;
  createdAt: string;
};

const chatsStore: Writable<Chat[]> = writable([]);
const selectedChatStore: Writable<string | null> = writable(null);

function createId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function createChat(title?: string) {
  const id = createId();
  const chat: Chat = {
    id,
    title: title ?? "New Chat",
    createdAt: new Date().toISOString(),
  };
  console.debug('createChat called', chat);
  chatsStore.update((c) => [chat, ...c]);
  selectedChatStore.set(id);
  return chat;
}

export function selectChat(id: string) {
  selectedChatStore.set(id);
}

export const chats = {
  subscribe: chatsStore.subscribe,
};

export const selectedChat = {
  subscribe: selectedChatStore.subscribe,
};

export default {
  createChat,
  selectChat,
  chats,
  selectedChat,
};
