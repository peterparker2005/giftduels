import { z } from "zod";

const EnvSchema = z.object({
	VITE_SENTRY_DSN: z.string().url(),
	VITE_LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),
	VITE_API_URL: z.string().url(),
});

const _env = EnvSchema.safeParse(import.meta.env);

if (!_env.success) {
	throw new Error("invalid environment variables");
}

type Env = z.infer<typeof EnvSchema>;

// Конфигурация как singleton с get'ерами
class Config {
	private readonly env: Env;

	constructor(env: Env) {
		this.env = env;
	}

	get isProd() {
		return import.meta.env.PROD;
	}

	get isDev() {
		return import.meta.env.DEV;
	}

	get logger() {
		return {
			level: this.env.VITE_LOG_LEVEL,
		};
	}

	get sentry() {
		return {
			dsn: this.env.VITE_SENTRY_DSN,
		};
	}

	get apiUrl() {
		return this.env.VITE_API_URL;
	}
}

export const config = new Config(_env.data);
