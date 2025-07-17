import type { StreamRequest } from "@giftduels/protobuf-js/giftduels/event/v1/public_service_pb";

/**
 * Простая очередь для bidi‑стрима (AsyncIterable<StreamRequest>)
 */
export function createRequestQueue() {
	type Puller = (r: IteratorResult<StreamRequest>) => void;

	let done = false;
	const pullers: Puller[] = [];
	const buffer: StreamRequest[] = [];

	return {
		push(req: StreamRequest) {
			if (done) return;
			if (pullers.length) {
				const pull = pullers.shift();
				if (!pull) return;
				pull({ value: req, done: false });
			} else {
				buffer.push(req);
			}
		},
		close() {
			done = true;
			while (pullers.length) {
				const pull = pullers.shift();
				if (!pull) return;
				pull({ value: undefined, done: true });
			}
		},
		async *[Symbol.asyncIterator](): AsyncGenerator<
			StreamRequest,
			void,
			unknown
		> {
			while (true) {
				if (buffer.length) {
					const req = buffer.shift();
					if (!req) return;
					yield req;
					continue;
				}
				if (done) return;
				const item: IteratorResult<StreamRequest> = await new Promise(
					(resolve) => pullers.push(resolve),
				);
				if (item.done) return;
				yield item.value;
			}
		},
	};
}
