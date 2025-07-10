import { Context } from "grammy";
import { InvoiceService } from "@/services/invoiceService";

export interface ExtendedContext extends Context {
	services: {
		invoice: InvoiceService;
	};
}
