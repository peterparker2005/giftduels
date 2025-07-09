import { DescMessage, fromBinary } from "@bufbuild/protobuf";

/**
 * Декодирует Protobuf-сообщение в инстанс T.
 * @param msg — AMQP-сообщение
 * @param schema — схема (DescMessage<T>)
 * @returns инстанс T
 */
export function decodeProtobufMessage<T>(msg: Buffer, schema: DescMessage): T {
	return fromBinary(schema, msg) as T;
}
