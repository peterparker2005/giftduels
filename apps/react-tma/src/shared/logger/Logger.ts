import * as Sentry from "@sentry/react";

type LogLevel = "info" | "warn" | "error" | "debug";

type LoggerSink = (
	level: LogLevel,
	message: string,
	// biome-ignore lint/suspicious/noExplicitAny: expected
	ctx?: any,
	name?: string,
) => void;

type LoggerOptions = {
	name: string;
	parent?: LoggerInstance;
	sink?: LoggerSink;
};

export interface LoggerInstance {
	child(name: string): LoggerInstance;
	// biome-ignore lint/suspicious/noExplicitAny: expected
	info(msg: string, ctx?: any): void;
	// biome-ignore lint/suspicious/noExplicitAny: expected
	warn(msg: string, ctx?: any): void;
	// biome-ignore lint/suspicious/noExplicitAny: expected
	error(msg: string, ctx?: any): void;
	// biome-ignore lint/suspicious/noExplicitAny: expected
	debug(msg: string, ctx?: any): void;
	readonly name: string;
}

function getTime() {
	const d = new Date();
	return (
		d.toLocaleTimeString("ru-RU", { hour12: false }) +
		"." +
		String(d.getMilliseconds()).padStart(3, "0")
	);
}

function getLevelStyle(level: LogLevel) {
	switch (level) {
		case "info":
			return "background:#17a2fa;color:#fff;border-radius:16px;padding:0px 6px;font-weight:bold";
		case "warn":
			return "background:#ffc107;color:#333;border-radius:16px;padding:0px 6px;font-weight:bold";
		case "error":
			return "background:#dc3545;color:#fff;border-radius:16px;padding:0px 6px;font-weight:bold";
		case "debug":
			return "background:#6c757d;color:#fff;border-radius:16px;padding:0px 6px;font-weight:bold";
		default:
			return "";
	}
}

function getNameStyle() {
	return "background:#20b26b;color:#fff;border-radius:16px;padding:0px 6px;font-weight:bold";
}

interface InternalLoggerInstance extends LoggerInstance {
	_sink?: LoggerSink;
}

export function createLogger({
	name,
	parent,
	sink,
}: LoggerOptions): LoggerInstance {
	const fullName = parent ? `${parent.name}/${name}` : name;
	const logSink = sink || (parent as InternalLoggerInstance)?._sink;

	function formatArgs(
		level: LogLevel,
		message: string,
		// biome-ignore lint/suspicious/noExplicitAny: expected
		context?: Record<string, any>,
	) {
		const time = getTime();
		return [
			`%c${level.toUpperCase()} ${time}%c %c${fullName}%c – %s`,
			getLevelStyle(level),
			"",
			getNameStyle(),
			"",
			message,
			context || "",
		];
	}

	const logger: InternalLoggerInstance = {
		name: fullName,
		child(childName: string) {
			return createLogger({ name: childName, parent: logger, sink: logSink });
		},
		info(message, context) {
			console.info(...formatArgs("info", message, context));
			logSink?.("info", message, context, fullName);
		},
		warn(message, context) {
			console.warn(...formatArgs("warn", message, context));
			logSink?.("warn", message, context, fullName);
		},
		error(message, context) {
			console.error(...formatArgs("error", message, context));
			logSink?.("error", message, context, fullName);
		},
		debug(message, context) {
			console.debug(...formatArgs("debug", message, context));
			logSink?.("debug", message, context, fullName);
		},
		_sink: logSink,
	};
	return logger;
}

// USAGE:

const sentrySink: LoggerSink = (level, message, ctx, name) => {
	if (level === "error") {
		Sentry.captureException(new Error(`[${name}] ${message}`), {
			level: "error",
			extra: ctx,
			tags: { logger: name },
		});
	} else if (level === "warn") {
		Sentry.captureMessage(`[${name}] ${message}`, {
			level: "warning",
			extra: ctx,
			tags: { logger: name },
		});
	}
};

export const logger = createLogger({ name: "App", sink: sentrySink });
