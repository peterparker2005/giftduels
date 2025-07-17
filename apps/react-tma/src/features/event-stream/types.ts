import type { DuelEvent } from "@giftduels/protobuf-js/giftduels/duel/v1/event_pb";

// Ключи событий, которые приходят из стрима
export type EventKey = "duelEvent";
export interface EventPayloadMap {
	duelEvent: DuelEvent;
	[key: string]: unknown; // для mitt
	[symbol: symbol]: unknown; // для mitt
}
