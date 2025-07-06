/* eslint-disable @typescript-eslint/no-explicit-any */
import { logger } from '@/logger'
import { giftduels } from '@giftduels/protobuf-ts'
import { Api } from 'telegram'

/**
 * Parser for Telegram API gift objects to protobuf types
 */

function createGiftTelegramId(
	value: string | number | bigint
): giftduels.shared.v1.GiftTelegramId {
	return giftduels.shared.v1.GiftTelegramId.create({
		value: typeof value === 'string' ? value : value.toString(),
	})
}

function createTelegramUserId(
	value: string | number | bigint
): giftduels.shared.v1.TelegramUserId {
	return giftduels.shared.v1.TelegramUserId.create({
		value: typeof value === 'string' ? value : value.toString(),
	})
}

function createTimestamp(unixTimestamp: number): Date {
	return new Date(unixTimestamp * 1000)
}

function parseAttributes(
	telegramAttributes?: Array<{
		type?: string
		name?: string
		rarity?: number
		[key: string]: any
	}>
): giftduels.gift.v1.GiftAttribute[] {
	if (!telegramAttributes || !Array.isArray(telegramAttributes)) {
		return []
	}

	return telegramAttributes.map(attr => {
		let attributeType =
			giftduels.gift.v1.GiftAttributeType.GIFT_ATTRIBUTE_TYPE_UNSPECIFIED

		// Map attribute types from Telegram to protobuf
		switch (attr.type?.toLowerCase()) {
			case 'model':
				attributeType =
					giftduels.gift.v1.GiftAttributeType.GIFT_ATTRIBUTE_TYPE_MODEL
				break
			case 'backdrop':
				attributeType =
					giftduels.gift.v1.GiftAttributeType.GIFT_ATTRIBUTE_TYPE_BACKDROP
				break
			case 'symbol':
				attributeType =
					giftduels.gift.v1.GiftAttributeType.GIFT_ATTRIBUTE_TYPE_SYMBOL
				break
			default:
				attributeType =
					giftduels.gift.v1.GiftAttributeType.GIFT_ATTRIBUTE_TYPE_UNSPECIFIED
		}

		return giftduels.gift.v1.GiftAttribute.create({
			type: attributeType,
			name: attr.name || '',
			rarity: attr.rarity || 0,
			description: attr.description || '',
		})
	})
}

function slugify(title: string): string {
	return title
		.toLowerCase()
		.replace(/[^a-z0-9\s-]/g, '')
		.replace(/\s+/g, '-')
		.trim()
}

// ===== MAIN PARSERS =====

/**
 * Parse SavedStarGift from getUserGifts() API call
 */
export function parseSavedStarGift(
	savedGift: Api.SavedStarGift,
	ownerTelegramId: number
): giftduels.gift.v1.Gift {
	const gift = savedGift.gift

	if (!gift) {
		throw new Error('SavedStarGift does not contain gift object')
	}

	// Handle different gift types
	let telegramGiftId: string
	let title: string
	let slug: string
	let attributes: giftduels.gift.v1.GiftAttribute[] = []

	if (gift.className === 'StarGiftUnique') {
		const uniqueGift = gift as Api.StarGiftUnique
		telegramGiftId = uniqueGift.id?.toString() || '0'
		title = uniqueGift.title || 'Unknown Gift'
		slug = uniqueGift.slug || slugify(title)
		attributes = parseAttributes(uniqueGift.attributes as any)
	} else if (gift.className === 'StarGift') {
		const regularGift = gift as Api.StarGift
		telegramGiftId = regularGift.id?.toString() || '0'
		title = (regularGift as any).title || 'Unknown Gift'
		slug = slugify(title)
		// Regular gifts might not have attributes
	} else {
		logger.warn(`Unknown gift type: ${(gift as any).className as string}`)
		telegramGiftId = '0'
		title = 'Unknown Gift'
		slug = 'unknown'
	}

	// Determine status from saved gift flags
	let status = giftduels.gift.v1.GiftStatus.GIFT_STATUS_OWNED
	if (savedGift.refunded) {
		status = giftduels.gift.v1.GiftStatus.GIFT_STATUS_WITHDRAWN // Treat refunded as withdrawn
	}

	const result: giftduels.gift.v1.Gift = giftduels.gift.v1.Gift.create({
		telegramGiftId: createGiftTelegramId(telegramGiftId),
		date: createTimestamp(savedGift.date || Math.floor(Date.now() / 1000)),
		ownerTelegramId: createTelegramUserId(ownerTelegramId),
		collectibleId: 0, // Not available in SavedStarGift
		telegramMessageId: savedGift.msgId || 0,
		title,
		slug,
		imageUrl: '', // Not available in SavedStarGift
		attributes,
		originalPrice: undefined, // Not available in SavedStarGift
		status,
		withdrawnAt: undefined,
	})

	logger.debug({ giftId: telegramGiftId, title }, '📦 Parsed SavedStarGift')

	return result
}

/**
 * Parse MessageActionStarGift for regular star gifts
 */
export function parseMessageActionStarGift(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number
): giftduels.gift.v1.TelegramGiftReceivedEvent {
	const action = message.action as Api.MessageActionStarGift

	if (!action.gift) {
		throw new Error('MessageActionStarGift does not contain gift object')
	}

	const gift = action.gift
	let telegramGiftId: string
	let title: string
	let slug: string
	const attributes: giftduels.gift.v1.GiftAttribute[] = []

	if (gift.className === 'StarGift') {
		const starGift = gift as Api.StarGift
		telegramGiftId = starGift.id?.toString() || '0'
		title = (starGift as any).title || 'Star Gift'
		slug = slugify(title)
	} else {
		logger.warn(
			`Unexpected gift type in MessageActionStarGift: ${gift.className}`
		)
		telegramGiftId = '0'
		title = 'Unknown Gift'
		slug = 'unknown'
	}

	const result: giftduels.gift.v1.TelegramGiftReceivedEvent =
		giftduels.gift.v1.TelegramGiftReceivedEvent.create({
			telegramGiftId: createGiftTelegramId(telegramGiftId),
			depositDate: createTimestamp(
				message.date || Math.floor(Date.now() / 1000)
			),
			ownerTelegramId: createTelegramUserId(toUserId || fromUserId),
			title,
			slug,
			attributes,
			collectibleId: 0,
			upgradeMessageId: 0,
		})

	logger.debug(
		{ giftId: telegramGiftId, title, fromUserId, toUserId },
		'🎁 Parsed MessageActionStarGift'
	)

	return result
}

/**
 * Parse MessageActionStarGiftUnique for unique/NFT star gifts
 */
export function parseMessageActionStarGiftUnique(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number
): giftduels.gift.v1.TelegramGiftReceivedEvent {
	const action = message.action as Api.MessageActionStarGiftUnique

	if (!action.gift) {
		throw new Error('MessageActionStarGiftUnique does not contain gift object')
	}

	const gift = action.gift
	let telegramGiftId: string
	let title: string
	let slug: string
	let attributes: giftduels.gift.v1.GiftAttribute[] = []
	let collectibleId = 0

	if (gift.className === 'StarGiftUnique') {
		const uniqueGift = gift as Api.StarGiftUnique
		telegramGiftId = uniqueGift.id?.toString() || '0'
		title = uniqueGift.title || 'Unique Gift'
		slug = uniqueGift.slug || slugify(title)
		attributes = parseAttributes(uniqueGift.attributes as any)
		collectibleId = uniqueGift.num || 0 // Use num as collectible ID
	} else {
		logger.warn(
			`Unexpected gift type in MessageActionStarGiftUnique: ${gift.className}`
		)
		telegramGiftId = '0'
		title = 'Unknown Unique Gift'
		slug = 'unknown'
	}

	const result: giftduels.gift.v1.TelegramGiftReceivedEvent =
		giftduels.gift.v1.TelegramGiftReceivedEvent.create({
			telegramGiftId: createGiftTelegramId(telegramGiftId),
			depositDate: createTimestamp(
				message.date || Math.floor(Date.now() / 1000)
			),
			ownerTelegramId: createTelegramUserId(toUserId || fromUserId),
			title,
			slug,
			attributes,
			collectibleId,
			upgradeMessageId: message.id || 0,
		})

	logger.debug(
		{
			giftId: telegramGiftId,
			title,
			collectibleId,
			fromUserId,
			toUserId,
		},
		'🎁 Parsed MessageActionStarGiftUnique'
	)

	return result
}

/**
 * Parse SavedStarGift to GiftView for API responses
 */
export function parseSavedStarGiftToView(
	savedGift: Api.SavedStarGift,
	ownerTelegramId: number
): giftduels.gift.v1.GiftView {
	const fullGift = parseSavedStarGift(savedGift, ownerTelegramId)

	return giftduels.gift.v1.GiftView.create({
		giftId: fullGift.giftId,
		telegramGiftId: fullGift.telegramGiftId,
		title: fullGift.title,
		slug: fullGift.slug,
		imageUrl: fullGift.imageUrl,
		originalPrice: fullGift.originalPrice,
		collectibleId: fullGift.collectibleId,
		status: fullGift.status,
		withdrawnAt: fullGift.withdrawnAt,
	})
}

/**
 * Helper function used in existing handler - exports parseNftGift alias
 */
export function parseNftGift(
	message: Api.MessageService,
	fromUserId: number,
	toUserId?: number
): giftduels.gift.v1.TelegramGiftReceivedEvent {
	if (message.action instanceof Api.MessageActionStarGiftUnique) {
		return parseMessageActionStarGiftUnique(message, fromUserId, toUserId)
	} else if (message.action instanceof Api.MessageActionStarGift) {
		return parseMessageActionStarGift(message, fromUserId, toUserId)
	} else {
		throw new Error(
			`Unsupported message action type: ${message.action.className}`
		)
	}
}

// ===== BATCH OPERATIONS =====

/**
 * Parse multiple SavedStarGifts to Gift objects
 */
export function parseSavedStarGifts(
	savedGifts: Api.SavedStarGift[],
	ownerTelegramId: number
): giftduels.gift.v1.Gift[] {
	return savedGifts.map(savedGift =>
		parseSavedStarGift(savedGift, ownerTelegramId)
	)
}
/**
 * Parse multiple SavedStarGifts to GiftView objects
 */
export function parseSavedStarGiftsToViews(
	savedGifts: Api.SavedStarGift[],
	ownerTelegramId: number
): giftduels.gift.v1.GiftView[] {
	return savedGifts.map(savedGift =>
		parseSavedStarGiftToView(savedGift, ownerTelegramId)
	)
}
