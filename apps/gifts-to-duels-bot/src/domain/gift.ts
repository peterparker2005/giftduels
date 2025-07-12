// biome-ignore-all lint/suspicious/noExplicitAny: all

import { create } from "@bufbuild/protobuf";
import { TimestampSchema } from "@bufbuild/protobuf/wkt";
import {
	TelegramGiftReceivedEvent,
	TelegramGiftReceivedEventSchema,
} from "@giftduels/protobuf-js/giftduels/gift/v1/events_pb";
import {
	Gift,
	GiftAttributeBackdrop,
	GiftAttributeBackdropSchema,
	GiftAttributeModel,
	GiftAttributeModelSchema,
	GiftAttributeSymbol,
	GiftAttributeSymbolSchema,
	GiftSchema,
	GiftStatus,
	GiftView,
	GiftViewSchema,
} from "@giftduels/protobuf-js/giftduels/gift/v1/gift_pb";
import {
	GiftTelegramId,
	GiftTelegramIdSchema,
	TelegramUserId,
	TelegramUserIdSchema,
} from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { Api } from "telegram";
import { logger } from "@/logger";

/**
 * Parser for Telegram API gift objects to protobuf types
 */

function createGiftTelegramId(value: string | number | bigint): GiftTelegramId {
	return create(GiftTelegramIdSchema, {
		value: BigInt(value),
	});
}

function createTelegramUserId(value: string | number | bigint): TelegramUserId {
	return create(TelegramUserIdSchema, {
		value: BigInt(value),
	});
}

function createTimestamp(unixTimestamp: number) {
	const seconds = Math.floor(unixTimestamp);
	const nanos = (unixTimestamp % 1) * 1e9;
	return create(TimestampSchema, { seconds: BigInt(seconds), nanos });
}

interface ParsedAttributes {
	backdrop?: GiftAttributeBackdrop;
	model?: GiftAttributeModel;
	symbol?: GiftAttributeSymbol;
}

function parseAttributes(
	telegramAttributes?: Api.TypeStarGiftAttribute[],
): ParsedAttributes {
	const result: ParsedAttributes = {};

	if (!telegramAttributes || !Array.isArray(telegramAttributes)) {
		return result;
	}

	telegramAttributes.forEach((attr) => {
		if (attr instanceof Api.StarGiftAttributeBackdrop) {
			result.backdrop = create(GiftAttributeBackdropSchema, {
				name: attr.name,
				rarityPerMille: attr.rarityPermille,
				centerColor: (attr as any).centerColor || "",
				edgeColor: (attr as any).edgeColor || "",
				patternColor: (attr as any).patternColor || "",
				textColor: (attr as any).textColor || "",
			});
		} else if (attr instanceof Api.StarGiftAttributeModel) {
			result.model = create(GiftAttributeModelSchema, {
				name: attr.name,
				rarityPerMille: attr.rarityPermille,
			});
		} else if (attr instanceof Api.StarGiftAttributePattern) {
			result.symbol = create(GiftAttributeSymbolSchema, {
				name: attr.name,
				rarityPerMille: attr.rarityPermille,
			});
		} else if (attr instanceof Api.StarGiftAttributeOriginalDetails) {
			logger.warn(
				"⚠️ StarGiftAttributeOriginalDetails is not handled yet, skipping",
			);
		} else {
			logger.warn(
				{ className: (attr as any)?.className },
				"⚠️ Unknown Telegram gift attribute type, skipping",
			);
		}
	});

	return result;
}

function slugify(title: string): string {
	return title
		.toLowerCase()
		.replace(/[^a-z0-9\s-]/g, "")
		.replace(/\s+/g, "-")
		.trim();
}

// ===== MAIN PARSERS =====

/**
 * Parse SavedStarGift from getUserGifts() API call
 */
export function parseSavedStarGift(
	savedGift: Api.SavedStarGift,
	ownerTelegramId: number,
): Gift {
	const gift = savedGift.gift;

	if (!gift) {
		throw new Error("SavedStarGift does not contain gift object");
	}

	// Handle different gift types
	let telegramGiftId: string;
	let title: string;
	let slug: string;
	let attributes: ParsedAttributes = {};
	let collectibleId = 0;

	if (gift.className === "StarGiftUnique") {
		const uniqueGift = gift as Api.StarGiftUnique;
		telegramGiftId = uniqueGift.id?.toString() || "0";
		title = uniqueGift.title || "Unique Gift";
		slug = uniqueGift.slug || slugify(title);
		attributes = parseAttributes(uniqueGift.attributes);
		collectibleId = uniqueGift.num;
	} else if (gift.className === "StarGift") {
		const regularGift = gift as Api.StarGift;
		telegramGiftId = regularGift.id?.toString() || "0";
		title = (regularGift as any).title || "Unknown Gift";
		slug = slugify(title);
		// Regular gifts might not have attributes
	} else {
		logger.warn(`Unknown gift type: ${(gift as any).className as string}`);
		telegramGiftId = "0";
		title = "Unknown Gift";
		slug = "unknown";
	}

	// Determine status from saved gift flags
	let status = GiftStatus.OWNED;
	if (savedGift.refunded) {
		status = GiftStatus.WITHDRAWN; // Treat refunded as withdrawn
	}

	const result: Gift = create(GiftSchema, {
		telegramGiftId: createGiftTelegramId(telegramGiftId),
		date: createTimestamp(savedGift.date || Math.floor(Date.now() / 1000)),
		ownerTelegramId: createTelegramUserId(ownerTelegramId),
		collectibleId,
		telegramMessageId: savedGift.msgId || 0,
		title,
		slug,
		backdrop: attributes.backdrop,
		model: attributes.model,
		symbol: attributes.symbol,
		status,
		withdrawnAt: undefined,
	});

	logger.debug({ giftId: telegramGiftId, title }, "📦 Parsed SavedStarGift");

	return result;
}

/**
 * Parse MessageActionStarGift for regular star gifts
 */
export function parseMessageActionStarGift(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number,
): TelegramGiftReceivedEvent {
	const action = message.action as Api.MessageActionStarGift;

	if (!action.gift) {
		throw new Error("MessageActionStarGift does not contain gift object");
	}

	const gift = action.gift;
	let telegramGiftId: string;
	let title: string;
	let slug: string;
	const attributes: ParsedAttributes = {};

	if (gift.className === "StarGift") {
		const starGift = gift as Api.StarGift;
		telegramGiftId = starGift.id?.toString() || "0";
		title = (starGift as any).title || "Star Gift";
		slug = slugify(title);
	} else {
		logger.warn(
			`Unexpected gift type in MessageActionStarGift: ${gift.className}`,
		);
		telegramGiftId = "0";
		title = "Unknown Gift";
		slug = "unknown";
	}

	const result: TelegramGiftReceivedEvent = create(
		TelegramGiftReceivedEventSchema,
		{
			telegramGiftId: createGiftTelegramId(telegramGiftId),
			depositDate: createTimestamp(
				message.date || Math.floor(Date.now() / 1000),
			),
			ownerTelegramId: createTelegramUserId(fromUserId),
			title,
			slug,
			backdrop: attributes.backdrop,
			model: attributes.model,
			symbol: attributes.symbol,
			collectibleId: 0,
			upgradeMessageId: 0,
		},
	);

	logger.debug(
		{ giftId: telegramGiftId, title, fromUserId, toUserId },
		"🎁 Parsed MessageActionStarGift",
	);

	return result;
}

/**
 * Parse MessageActionStarGiftUnique for unique/NFT star gifts
 */
export function parseMessageActionStarGiftUnique(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number,
): TelegramGiftReceivedEvent {
	const action = message.action as Api.MessageActionStarGiftUnique;

	if (!action.gift) {
		throw new Error("MessageActionStarGiftUnique does not contain gift object");
	}

	const gift = action.gift;
	let telegramGiftId: string;
	let title: string;
	let slug: string;
	let attributes: ParsedAttributes = {};
	let collectibleId = 0;

	if (gift.className === "StarGiftUnique") {
		const uniqueGift = gift as Api.StarGiftUnique;
		telegramGiftId = uniqueGift.id?.toString() || "0";
		title = uniqueGift.title || "Unique Gift";
		slug = uniqueGift.slug || slugify(title);
		attributes = parseAttributes(uniqueGift.attributes);
		collectibleId = uniqueGift.num;
	} else {
		logger.warn(
			`Unexpected gift type in MessageActionStarGiftUnique: ${gift.className}`,
		);
		telegramGiftId = "0";
		title = "Unknown Unique Gift";
		slug = "unknown";
	}

	const result: TelegramGiftReceivedEvent = create(
		TelegramGiftReceivedEventSchema,
		{
			telegramGiftId: createGiftTelegramId(telegramGiftId),
			depositDate: createTimestamp(
				message.date || Math.floor(Date.now() / 1000),
			),
			ownerTelegramId: createTelegramUserId(fromUserId),
			title,
			slug,
			backdrop: attributes.backdrop,
			model: attributes.model,
			symbol: attributes.symbol,
			collectibleId,
			upgradeMessageId: message.id || 0,
		},
	);

	logger.debug(
		{
			giftId: telegramGiftId,
			title,
			collectibleId,
			fromUserId,
			toUserId,
		},
		"🎁 Parsed MessageActionStarGiftUnique",
	);

	return result;
}

/**
 * Parse SavedStarGift to GiftView for API responses
 */
export function parseSavedStarGiftToView(
	savedGift: Api.SavedStarGift,
	ownerTelegramId: number,
): GiftView {
	const fullGift = parseSavedStarGift(savedGift, ownerTelegramId);

	return create(GiftViewSchema, {
		giftId: fullGift.giftId,
		telegramGiftId: fullGift.telegramGiftId,
		title: fullGift.title,
		slug: fullGift.slug,
		price: fullGift.price,
		collectibleId: fullGift.collectibleId,
		status: fullGift.status,
		withdrawnAt: fullGift.withdrawnAt,
		backdrop: fullGift.backdrop,
		model: fullGift.model,
		symbol: fullGift.symbol,
	});
}

/**
 * Helper function used in existing handler - exports parseNftGift alias
 */
export function parseNftGift(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number,
): TelegramGiftReceivedEvent {
	if (message.action instanceof Api.MessageActionStarGiftUnique) {
		return parseMessageActionStarGiftUnique(message, fromUserId, toUserId);
	}

	throw new Error(
		`Unsupported message action type: ${message.action.className}`,
	);
}

/**
 * Parse SavedStarGift to TelegramGiftReceivedEvent for event processing
 */
export function parseSavedStarGiftToEvent(
	savedGift: Api.SavedStarGift,
	ownerTelegramId: number,
): TelegramGiftReceivedEvent {
	const gift = savedGift.gift;

	if (!gift) {
		throw new Error("SavedStarGift does not contain gift object");
	}

	// Handle different gift types
	let telegramGiftId: string;
	let title: string;
	let slug: string;
	let attributes: ParsedAttributes = {};
	let collectibleId = 0;

	if (gift.className === "StarGiftUnique") {
		const uniqueGift = gift as Api.StarGiftUnique;
		telegramGiftId = uniqueGift.id?.toString() || "0";
		title = uniqueGift.title || "Unique Gift";
		slug = uniqueGift.slug || slugify(title);
		attributes = parseAttributes(uniqueGift.attributes);
		collectibleId = uniqueGift.num;
	} else {
		logger.warn(`Unknown gift type: ${(gift as any).className as string}`);
		telegramGiftId = "0";
		title = "Unknown Gift";
		slug = "unknown";
		collectibleId = 0;
	}

	const result: TelegramGiftReceivedEvent = create(
		TelegramGiftReceivedEventSchema,
		{
			telegramGiftId: createGiftTelegramId(telegramGiftId),
			depositDate: createTimestamp(
				savedGift.date || Math.floor(Date.now() / 1000),
			),
			ownerTelegramId: createTelegramUserId(ownerTelegramId),
			title,
			slug,
			backdrop: attributes.backdrop,
			model: attributes.model,
			symbol: attributes.symbol,
			collectibleId,
			upgradeMessageId: savedGift.msgId || 0,
		},
	);

	logger.debug(
		{
			giftId: telegramGiftId,
			title,
			collectibleId,
			ownerTelegramId,
		},
		"🎁 Parsed SavedStarGift to TelegramGiftReceivedEvent",
	);

	return result;
}

// ===== BATCH OPERATIONS =====

/**
 * Parse multiple SavedStarGifts to Gift objects
 */
export function parseSavedStarGifts(
	savedGifts: Api.SavedStarGift[],
	ownerTelegramId: number,
): Gift[] {
	return savedGifts.map((savedGift) =>
		parseSavedStarGift(savedGift, ownerTelegramId),
	);
}
/**
 * Parse multiple SavedStarGifts to GiftView objects
 */
export function parseSavedStarGiftsToViews(
	savedGifts: Api.SavedStarGift[],
	ownerTelegramId: number,
): GiftView[] {
	return savedGifts.map((savedGift) =>
		parseSavedStarGiftToView(savedGift, ownerTelegramId),
	);
}

/**
 * Parse multiple SavedStarGifts to TelegramGiftReceivedEvent objects
 */
export function parseSavedStarGiftsToEvents(
	savedGifts: Api.SavedStarGift[],
	ownerTelegramId: number,
): TelegramGiftReceivedEvent[] {
	return savedGifts.map((savedGift) =>
		parseSavedStarGiftToEvent(savedGift, ownerTelegramId),
	);
}
