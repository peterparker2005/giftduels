import {
	CreateStarInvoiceRequest,
	CreateStarInvoiceResponse,
} from "@giftduels/protobuf-js/giftduels/telegrambot/v1/private_service_pb";
import type {
	ServerUnaryCall,
	sendUnaryData,
	UntypedHandleCall,
} from "@grpc/grpc-js";
import { Status } from "@grpc/grpc-js/build/src/constants";
import { InvoiceService } from "../../services/invoiceService";

export function makeTelegramBotHandlers(
	invoiceService: InvoiceService,
): Record<string, UntypedHandleCall> {
	return {
		CreateStarInvoice(
			call: ServerUnaryCall<
				CreateStarInvoiceRequest,
				CreateStarInvoiceResponse
			>,
			callback: sendUnaryData<CreateStarInvoiceResponse>,
		) {
			const req = call.request;
			invoiceService
				.createStarInvoice({
					starsAmount: req.starsAmount?.value ?? 0,
					title: req.title,
					description: req.description,
					payload: req.payload,
				})
				.then((invoiceUrl) => {
					callback(null, {
						invoiceUrl,
						$typeName: "giftduels.telegrambot.v1.CreateStarInvoiceResponse",
					});
				})
				.catch((err) => {
					callback({ code: Status.INTERNAL, message: err.message }, null);
				});
		},
	};
}
