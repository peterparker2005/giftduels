import { useCallback, useEffect, useMemo } from "react";
import { streamManager } from "../StreamManager";
import type { EventKey, EventPayloadMap } from "../types";

export function useStream() {
	useEffect(() => {
		streamManager.start();
		return () => {
			streamManager.stop();
		};
	}, []);

	const handle = useCallback(streamManager.on.bind(streamManager), []);
	const isConnected = useMemo(() => streamManager.isConnected(), []);

	return {
		handle,
		isConnected,
	};
}

/**
 * UseStreamHandler is a hook for automatically managing subscriptions to events
 * Automatically subscribes when mounted and unsubscribes when unmounted
 */
export function useStreamHandler<E extends EventKey>(
	eventType: E,
	handler: (payload: EventPayloadMap[E]) => void,
) {
	const { handle } = useStream();

	useEffect(() => {
		const unsubscribe = handle(eventType, handler);
		return unsubscribe;
	}, [handle, eventType, handler]);
}
