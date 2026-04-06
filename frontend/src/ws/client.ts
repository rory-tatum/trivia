import type { IncomingMessage, OutgoingMessage } from "./messages";
import { RECONNECT_FAILED } from "./events";

// Exponential backoff configuration
const BACKOFF_BASE_MS = 1000;
const BACKOFF_MULTIPLIER = 2;
const BACKOFF_MAX_MS = 30_000;
const MAX_RECONNECT_ATTEMPTS = 10;

export type MessageHandler = (msg: OutgoingMessage) => void;
export type EventHandler = (event: string) => void;

export class WsClient {
  private socket: WebSocket | null = null;
  private attempt = 0;
  private backoffMs = BACKOFF_BASE_MS;
  private closed = false;

  private messageHandlers: MessageHandler[] = [];
  private eventHandlers: Map<string, EventHandler[]> = new Map();

  constructor(private readonly url: string) {}

  connect(): void {
    this.closed = false;
    this.attempt = 0;
    this.backoffMs = BACKOFF_BASE_MS;
    this.openSocket();
  }

  disconnect(): void {
    this.closed = true;
    this.socket?.close();
    this.socket = null;
  }

  send(msg: IncomingMessage): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(msg));
    }
  }

  onMessage(handler: MessageHandler): void {
    this.messageHandlers.push(handler);
  }

  on(event: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event) ?? [];
    handlers.push(handler);
    this.eventHandlers.set(event, handlers);
  }

  private openSocket(): void {
    const ws = new WebSocket(this.url);
    this.socket = ws;

    ws.onopen = () => {
      this.attempt = 0;
      this.backoffMs = BACKOFF_BASE_MS;
    };

    ws.onmessage = (ev: MessageEvent) => {
      try {
        const msg = JSON.parse(ev.data as string) as OutgoingMessage;
        for (const handler of this.messageHandlers) {
          handler(msg);
        }
      } catch {
        // Ignore unparseable frames
      }
    };

    ws.onclose = () => {
      if (this.closed) return;
      this.scheduleReconnect();
    };

    ws.onerror = () => {
      ws.close();
    };
  }

  private scheduleReconnect(): void {
    if (this.attempt >= MAX_RECONNECT_ATTEMPTS) {
      this.emit(RECONNECT_FAILED);
      return;
    }
    const delay = Math.min(this.backoffMs, BACKOFF_MAX_MS);
    setTimeout(() => {
      if (this.closed) return;
      this.attempt += 1;
      this.backoffMs = Math.min(
        this.backoffMs * BACKOFF_MULTIPLIER,
        BACKOFF_MAX_MS,
      );
      this.openSocket();
    }, delay);
  }

  private emit(event: string): void {
    const handlers = this.eventHandlers.get(event) ?? [];
    for (const handler of handlers) {
      handler(event);
    }
  }
}
