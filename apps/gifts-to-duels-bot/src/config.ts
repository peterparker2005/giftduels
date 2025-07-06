import dotenv from 'dotenv'
import { z } from 'zod'

dotenv.config()

const EnvSchema = z.object({
	NODE_ENV: z
		.enum(['development', 'production', 'test'])
		.default('development'),

	// Telegram
	API_ID: z.string().transform(Number),
	API_HASH: z.string(),
	SESSION_STRING: z.string(),

	// Amqp
	AMQP_HOST: z.string(),
	AMQP_PORT: z.coerce.number().default(5672),
	AMQP_USER: z.string(),
	AMQP_PASSWORD: z.string(),
	AMQP_VHOST: z.string().default('/'),

	// S3
	S3_ENDPOINT: z.string().url(),
	S3_ACCESS_KEY: z.string(),
	S3_SECRET_KEY: z.string(),
	S3_BUCKET: z.string(),
	S3_REGION: z.string(),

	// Logging
	LOG_LEVEL: z.enum(['debug', 'info', 'warn', 'error']).default('info'),
})

const _env = EnvSchema.safeParse(process.env)

if (!_env.success) {
	console.error('❌ Invalid environment variables:', _env.error.format())
	process.exit(1)
}

type Env = z.infer<typeof EnvSchema>

// Конфигурация как singleton с get'ерами
class Config {
	private readonly env: Env

	constructor(env: Env) {
		this.env = env
	}

	get isProd() {
		return this.env.NODE_ENV === 'production'
	}

	get isDev() {
		return this.env.NODE_ENV === 'development'
	}

	get telegram() {
		return {
			apiId: this.env.API_ID,
			apiHash: this.env.API_HASH,
			sessionString: this.env.SESSION_STRING,
		}
	}

	get amqp() {
		return {
			host: this.env.AMQP_HOST,
			port: this.env.AMQP_PORT,
			user: this.env.AMQP_USER,
			password: this.env.AMQP_PASSWORD,
			vhost: this.env.AMQP_VHOST,
			url: () =>
				`amqp://${this.env.AMQP_USER}:${this.env.AMQP_PASSWORD}@${this.env.AMQP_HOST}:${this.env.AMQP_PORT}/${this.env.AMQP_VHOST}`,
		}
	}

	get s3() {
		return {
			endpoint: this.env.S3_ENDPOINT,
			accessKey: this.env.S3_ACCESS_KEY,
			secretKey: this.env.S3_SECRET_KEY,
			bucket: this.env.S3_BUCKET,
			region: this.env.S3_REGION,
		}
	}

	get logLevel() {
		return this.env.LOG_LEVEL
	}
}

export const config = new Config(_env.data)
