import { Composer } from "grammy";
import type { ExtendedContext } from "@/types/context";

export const paymentsRouter = new Composer<ExtendedContext>();

// 1) Pre-checkout: telegram присылает запрос, надо ответить за 10s
paymentsRouter.on("pre_checkout_query", async (ctx, next) => {
	// const q = ctx.update.pre_checkout_query;
	// const payload = q.invoice_payload; // base64 из CreateInvoiceRequest
	const ok = await ctx.services.invoice.handlePreCheckout();
	if (ok) {
		await ctx.answerPreCheckoutQuery(true);
		return next(); // дальше telegram пошлёт successful_payment
	} else {
		await ctx.answerPreCheckoutQuery(false, "Недостаточно звёзд на балансе");
	}
});

// 2) Successful payment: пользователь подтвердил и заплатил
paymentsRouter.on("message:successful_payment", async (ctx) => {
	const p = ctx.update.message.successful_payment;
	const payload = p.invoice_payload;
	const invoiceId = p.provider_payment_charge_id;
	const starsAmount = Number(p.total_amount) / 100; // total_amount в центах
	// публикуем событие в AMQP
	await ctx.services.invoice.handleSuccessfulPayment(
		ctx.from.id,
		payload,
		invoiceId,
		starsAmount,
	);
	// отправляем пользователю подтверждение
	await ctx.reply(`✅ Платёж на ${starsAmount} ⭐ успешно проведён!`);
});
