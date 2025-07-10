import * as path from "node:path";
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";

const PROTO_ROOT = path.resolve(__dirname, "../../../../packages/protobuf/api");
const PROTO_PATH = path.join(
	PROTO_ROOT,
	"giftduels/telegrambot/v1/private_service.proto",
);

export function loadTelegramBotService() {
	const packageDef = protoLoader.loadSync(PROTO_PATH, {
		keepCase: false,
		longs: String,
		enums: String,
		defaults: true,
		oneofs: true,
		includeDirs: [PROTO_ROOT],
	});
	// biome-ignore lint/suspicious/noExplicitAny: expected. fuckin node grpc loader
	const grpcObj = grpc.loadPackageDefinition(packageDef) as any;
	const serviceDef =
		grpcObj.giftduels.telegrambot.v1.TelegramBotPrivateService.service;
	return { packageDef, serviceDef };
}
