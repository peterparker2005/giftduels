import { createLogger } from '@giftduels/logger-ts'
import { config } from './config'

export const logger = createLogger({
	env: config.isDev ? 'development' : 'production',
	level: config.logLevel,
})
