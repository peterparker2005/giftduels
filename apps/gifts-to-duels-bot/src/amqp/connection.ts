import { connect, AmqpConnectionManagerOptions } from 'amqp-connection-manager'
import { logger } from '@/logger'
import { config } from '@/config'

type AmqpConn = ReturnType<typeof connect>

let _connection: AmqpConn | undefined
let resolveConn!: (c: AmqpConn) => void

/** Промис, который резолвится после connectAmqp() */
const connectionReady: Promise<AmqpConn> = new Promise(res => {
	resolveConn = res
})

export async function connectAmqp() {
	const amqpUrl = config.amqp.url()
	logger.info(
		`[AMQP] Attempting to connect to: ${amqpUrl.replace(/:.*@/, ':***@')}`
	)

	try {
		_connection = connect([amqpUrl], {
			heartbeatIntervalInSeconds: 15,
			reconnectTimeInSeconds: 5,
		} satisfies AmqpConnectionManagerOptions)

		_connection.on('connect', () => logger.info('[AMQP] connected'))
		_connection.on('disconnect', ({ err }) =>
			logger.warn({ err }, '[AMQP] disconnected')
		)
		_connection.on('error', err =>
			logger.error({ err }, '[AMQP] connection error')
		)

		resolveConn(_connection) // 🔔 соединение готово
		logger.info('[AMQP] Connection manager created successfully')
	} catch (err) {
		logger.error({ err }, '[AMQP] Failed to create connection')
		throw err
	}
}

/** Получить connection или дождаться, пока он появится */
export async function getConnection(): Promise<AmqpConn> {
	return _connection ?? connectionReady
}

export async function closeAmqp() {
	if (_connection) await _connection.close()
}
