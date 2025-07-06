import pino, { type Logger, type LoggerOptions, type Bindings } from 'pino'

export interface LoggerConfig {
	env: 'development' | 'production'
	level?: string
}

export function createLogger(config: LoggerConfig): Logger {
	const isProd = config.env === 'production'

	const baseOptions: LoggerOptions = {
		level: config.level ?? (isProd ? 'info' : 'debug'),
		timestamp: pino.stdTimeFunctions.isoTime,
	}

	return isProd
		? pino(baseOptions)
		: pino({
				...baseOptions,
				transport: {
					target: 'pino-pretty',
					options: {
						colorize: true,
						translateTime: 'HH:MM:ss',
						ignore: 'pid,hostname',
					},
				},
			})
}

export function createStream(logger: Logger) {
	return {
		write: (msg: string) => logger.info(msg.trimEnd()),
	}
}

export function createChildLogger(logger: Logger, bindings: Bindings): Logger {
	return logger.child(bindings)
}
