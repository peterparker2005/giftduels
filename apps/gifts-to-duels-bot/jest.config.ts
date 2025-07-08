import type { Config } from "jest";

const config: Config = {
	preset: "ts-jest",
	testEnvironment: "node",

	// оставляем только .ts, .js и без указания .js — он будет ESM по package.json
	extensionsToTreatAsEsm: [".ts"],

	// transform всех .ts и .js через ts-jest в режиме ESM
	transform: {
		"^.+\\.[tj]sx?$": [
			"ts-jest",
			{
				useESM: true,
				tsconfig: "tsconfig.json",
				// отключить лишние варнинги резолва .js
				diagnostics: false,
			},
		],
	},

	testMatch: ["**/__tests__/**/*.test.ts"],
	moduleFileExtensions: ["ts", "js", "json"],

	// не игнорируем модули telegram, @bufbuild и @gice
	transformIgnorePatterns: ["node_modules/(?!(telegram|@bufbuild|@gice)/)"],

	moduleNameMapper: {
		// фикс глубоких импортов gift_pb.js → .ts
		"^(\\.{1,2}/.*)\\.js$": "$1",
		"^@gice/protobuf/(.*)$":
			"<rootDir>/../../packages/protobuf/js/dist/$1/index.js",
		"^@/(.*)$": "<rootDir>/src/$1",
	},
};

export default config;
