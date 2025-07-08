const { build } = require("esbuild");
const path = require("node:path");

build({
	entryPoints: [path.resolve(__dirname, "../src/index.ts")],
	outfile: path.resolve(__dirname, "../dist/index.cjs"),
	bundle: true,
	platform: "node",
	format: "cjs",
	target: "node20",
	sourcemap: true,
	external: [
		"amqplib",
		"pino",
		"big-integer",
		"telegram",
		"@grpc/grpc-js",
		"@grpc/proto-loader",
	],
}).catch(() => process.exit(1));
