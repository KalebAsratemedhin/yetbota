/** OpenAPI types for ai-service RAG chat */

export type ChatRole = "user" | "assistant";

export type ChatTurn = {
  role: ChatRole;
  content: string;
};

export type ChatRequest = {
  message: string;
  history?: ChatTurn[];
};

export type ChatSource = {
  kind: string;
  entity_id: string;
  score: number;
  title?: string;
  text?: string;
  parent_id?: string | null;
  post_id?: string | null;
};

export type ChatResponse = {
  reply: string;
  model: string;
  sources: ChatSource[];
};
