import { Api, TelegramClient } from 'telegram'

import { parseMessageActionStarGiftUnique } from '@/domain/gift'
import { logger } from '@/logger'
import { publishProto } from '@/amqp/publisher'
import { giftduels } from '@giftduels/protobuf-ts'
import { encoderFor } from '@/utils/encoderFor'

export async function nftGiftHandler(client: TelegramClient) {
	client.addEventHandler(async (update: Api.TypeUpdate) => {
		if (!(update instanceof Api.UpdateNewMessage)) return

		const message = update.message
		logger.info({ messageType: message.className }, '📨 Incoming message')

		if (!(message instanceof Api.MessageService)) return

		logger.info({ action: message.action.className }, 'action.className')

		// Обрабатываем только NFT подарки (MessageActionStarGiftUnique)
		if (!(message.action instanceof Api.MessageActionStarGiftUnique)) return

		logger.info({ action: message.action.className }, '🎯 NFT Gift action')
		logger.info({ data: message }, 'Data')

		let senderId: number
		const peer = message.fromId ?? message.peerId

		if ('userId' in peer) {
			senderId = peer.userId.toJSNumber?.()
		} else if ('chatId' in peer) {
			senderId = peer.chatId.toJSNumber?.()
		} else {
			logger.warn({ peer }, '⚠️ Unknown peer type')
			return
		}

		logger.info({ action: message.action.className }, 'Processing NFT Gift...')

		const self = await client.getMe()

		logger.info({ senderId }, '🎁 Got NFT Gift')

		try {
			const gift = parseMessageActionStarGiftUnique(
				message,
				senderId,
				self.id?.toJSNumber()
			)

			logger.debug({ gift }, '📦 Parsed NFT gift')

			await publishProto<giftduels.gift.v1.TelegramGiftReceivedEvent>({
				routingKey: 'telegram.gift.received',
				msg: gift,
				encoder: encoderFor(giftduels.gift.v1.TelegramGiftReceivedEvent),
			})

			logger.info(
				{
					messageId: message.id,
					userId: senderId,
					giftId: gift.telegramGiftId,
				},
				'📤 NFT Gift event published'
			)

			await client.sendMessage(senderId, {
				message: `🎁 ${
					gift.title || 'Подарок'
				} успешно добавлен в ваш профиль GiftDuels!\n\nИщите игру или создайте свою — @GiftDuels`,
			})
		} catch (err) {
			logger.error(
				{
					err,
					messageId: message.id,
					userId: senderId,
				},
				'❌ Error in NFT Gift handler'
			)

			await client.sendMessage(senderId, {
				message: '❌ Не удалось обработать подарок. Попробуйте позже.',
			})
		}
	})
}
