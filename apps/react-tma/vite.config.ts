import fs from "node:fs";
import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react-swc";
import { defineConfig, type ServerOptions } from "vite";
import svgr from "vite-plugin-svgr";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
	const isProd = mode === "production";

	const plugins = [
		react(),
		tailwindcss(),
		svgr({
			svgrOptions: {
				exportType: "default",
			},
		}),
	];

	const resolveConfig = {
		alias: {
			"@": path.resolve(__dirname, "./src"),
		},
	};

	const tailscaleServerConfig: ServerOptions = {
		https: {
			key: fs.readFileSync(
				"/Users/szn/Library/Containers/io.tailscale.ipn.macos/Data/setter.alpaca-wahoo.ts.net.key",
			),
			cert: fs.readFileSync(
				"/Users/szn/Library/Containers/io.tailscale.ipn.macos/Data/setter.alpaca-wahoo.ts.net.crt",
			),
		},
		hmr: {
			host: "setter.alpaca-wahoo.ts.net",
			port: 3443,
		},
		host: "0.0.0.0",
		port: 3443,
		strictPort: true,
	};

	return {
		test: {
			globals: true,
			environment: "jsdom",
		},
		plugins,
		resolve: resolveConfig,
		server: !isProd ? tailscaleServerConfig : undefined,
		preview: !isProd
			? {
					https: tailscaleServerConfig.https,
					host: tailscaleServerConfig.host,
					port: tailscaleServerConfig.port,
					strictPort: tailscaleServerConfig.strictPort,
				}
			: undefined,
	};
});
