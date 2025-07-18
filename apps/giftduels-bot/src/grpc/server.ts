import * as grpc from "@grpc/grpc-js";
import { getContainer } from "@/container";
import { loadTelegramBotService } from "./loader";
import { enableReflection } from "./reflection";
import { makeTelegramBotHandlers } from "./services/telegramBotPrivate";

export function createGrpcServer(): grpc.Server {
	const container = getContainer();
	const invoiceService = container.resolve("invoiceService");
	const duelService = container.resolve("duelService");
	const server = new grpc.Server();
	const { packageDef, serviceDef } = loadTelegramBotService();
	server.addService(
		serviceDef,
		makeTelegramBotHandlers(invoiceService, duelService),
	);
	enableReflection(server, packageDef);
	return server;
}
