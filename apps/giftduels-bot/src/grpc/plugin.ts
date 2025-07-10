import { ServerCredentials } from "@grpc/grpc-js";
import type { FastifyInstance } from "fastify";
import fp from "fastify-plugin";
import { config } from "../config";
import { createGrpcServer } from "./server";

declare module "fastify" {
	interface FastifyInstance {
		grpcServer: { start(): void };
	}
}

export const grpcServerPlugin = fp(async (fastify: FastifyInstance) => {
	const server = createGrpcServer();

	fastify.decorate("grpcServer", {
		start() {
			server.bindAsync(
				`${config.grpc.host}:${config.grpc.port}`,
				ServerCredentials.createInsecure(),
				(err, port) => {
					if (err) {
						fastify.log.error({ err }, "gRPC bind error");
						return;
					}
					fastify.log.info(`gRPC server listening on ${port}`);
				},
			);
		},
	});

	// гарантируем graceful shutdown
	fastify.addHook("onClose", (_instance, done) => {
		server.tryShutdown((err) => {
			if (err) fastify.log.error(err, "gRPC shutdown error");
			else fastify.log.info("gRPC server shut down");
			done();
		});
	});
});
