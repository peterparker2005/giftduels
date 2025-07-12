import { create } from "@bufbuild/protobuf";
import { Timestamp, TimestampSchema } from "@bufbuild/protobuf/wkt";

export function dateToProto(date: Date): Timestamp {
	return create(TimestampSchema, {
		seconds: BigInt(Math.floor(date.getTime() / 1000)),
		nanos: (date.getTime() % 1000) * 1000000,
	});
}
