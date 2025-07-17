import { Code, ConnectError } from "@connectrpc/connect";
import mitt, { Emitter, Handler } from "mitt";
import { eventClient } from "@/shared/api/client";
import { logger } from "@/shared/logger";
import type { EventKey, EventPayloadMap } from "./types";

type StreamEmitter = Emitter<EventPayloadMap>;

export class StreamManager {
	private emitter: StreamEmitter = mitt();
	private started = false;
	private abortCtrl?: AbortController;
	private reconnectTimeout?: number;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000; // Start with 1 second

	start(): void {
		if (this.started) return;
		this.started = true;
		this.reconnectAttempts = 0;
		this.reconnectDelay = 1000;

		this.connect();
	}

	private connect(): void {
		if (this.abortCtrl) {
			this.abortCtrl.abort();
		}

		this.abortCtrl = new AbortController();
		const responses = eventClient.stream(
			{}, // пустой request, т.к. мы перешли на server-stream
			{ signal: this.abortCtrl.signal },
		);

		(async () => {
			try {
				logger.info("[StreamManager] Starting stream connection...");
				for await (const res of responses) {
					if (res.payload?.case === "duelEvent") {
						this.emitter.emit("duelEvent", res.payload.value);
					}
				}
			} catch (err) {
				if (err instanceof ConnectError && err.code === Code.Canceled) {
					logger.error("[StreamManager] Stream connection canceled", err);
					return; // молча выходим
				}
				logger.error("[StreamManager] stream error", err);

				// Attempt to reconnect if we haven't exceeded max attempts
				if (
					this.started &&
					this.reconnectAttempts < this.maxReconnectAttempts
				) {
					this.scheduleReconnect();
				} else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
					logger.error("[StreamManager] Max reconnection attempts reached");
					this.started = false;
				}
			}
		})();
	}

	private scheduleReconnect(): void {
		this.reconnectAttempts++;
		const delay = Math.min(
			this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
			30000,
		); // Max 30 seconds

		logger.info(
			`[StreamManager] Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`,
		);

		this.reconnectTimeout = setTimeout(() => {
			if (this.started) {
				this.connect();
			}
		}, delay);
	}

	stop(): void {
		this.started = false;
		this.reconnectAttempts = 0;

		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = undefined;
		}

		this.abortCtrl?.abort();
	}

	on<E extends EventKey>(
		type: E,
		handler: Handler<EventPayloadMap[E]>,
	): () => void {
		this.emitter.on(type, handler);
		return () => {
			this.emitter.off(type, handler);
		};
	}

	isConnected(): boolean {
		return this.started && !this.abortCtrl?.signal.aborted;
	}
}

export const streamManager = new StreamManager();
