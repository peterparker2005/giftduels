import * as grpc from "@grpc/grpc-js";
import { PackageDefinition } from "@grpc/proto-loader";
import { ReflectionService } from "@grpc/reflection";

export function enableReflection(
	server: grpc.Server,
	packageDef: PackageDefinition,
) {
	const refl = new ReflectionService(packageDef);
	refl.addToServer(server);
}
