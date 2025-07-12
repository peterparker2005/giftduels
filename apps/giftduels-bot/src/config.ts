import dotenv from "dotenv";
import { z } from "zod";

dotenv.config();

const EnvSchema = z.object({
	NODE_ENV: z
		.enum(["development", "production", "test"])
		.default("development"),

	SERVICE_NAME: z.string().default("giftduels-bot"),
	// Telegram
	TELEGRAM_BOT_TOKEN: z.string(),
	TELEGRAM_WEB_APP_URL: z.url(),
	TELEGRAM_ADMIN_IDS: z.string().transform((ids) => ids.split(",").map(Number)),
	TELEGRAM_DUEL_CHANNEL_ID: z.string(),

	// Amqp
	AMQP_HOST: z.string(),
	AMQP_PORT: z.coerce.number().default(5672),
	AMQP_USER: z.string(),
	AMQP_PASSWORD: z.string(),
	AMQP_VHOST: z.string().default("/"),

	// Logging
	LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),

	// gRPC
	GRPC_TELEGRAM_BOT_SERVICE_HOST: z.string().default("0.0.0.0"),
	GRPC_TELEGRAM_BOT_SERVICE_PORT: z.coerce.number().default(50060),
});

const _env = EnvSchema.safeParse(process.env);

if (!_env.success) {
	console.error("❌ Invalid environment variables:", _env.error.format());
	process.exit(1);
}

type Env = z.infer<typeof EnvSchema>;

// Конфигурация как singleton с get'ерами
class Config {
	private readonly env: Env;

	constructor(env: Env) {
		this.env = env;
	}

	get isProd() {
		return this.env.NODE_ENV === "production";
	}

	get isDev() {
		return this.env.NODE_ENV === "development";
	}

	get serviceName() {
		return this.env.SERVICE_NAME;
	}

	get telegram() {
		return {
			botToken: this.env.TELEGRAM_BOT_TOKEN,
			webAppUrl: this.env.TELEGRAM_WEB_APP_URL,
			adminIds: this.env.TELEGRAM_ADMIN_IDS,
			duelChannelId: this.env.TELEGRAM_DUEL_CHANNEL_ID,
		};
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
		};
	}

	get logLevel() {
		return this.env.LOG_LEVEL;
	}

	get grpc() {
		return {
			host: this.env.GRPC_TELEGRAM_BOT_SERVICE_HOST,
			port: this.env.GRPC_TELEGRAM_BOT_SERVICE_PORT,
		};
	}
}

export const config = new Config(_env.data);
