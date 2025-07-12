import {
	AwilixContainer,
	asClass,
	asFunction,
	createContainer,
	InjectionMode,
} from "awilix";
import { Bot } from "grammy";
import { createBot } from "./bot";
import { DuelService } from "./services/duelService";
import { InvoiceService } from "./services/invoiceService";
import { NotificationService } from "./services/notification";
import { ExtendedContext } from "./types/context";

// Типизация контейнера
export interface Cradle {
	bot: Bot<ExtendedContext>;
	invoiceService: InvoiceService;
	notificationService: NotificationService;
	duelService: DuelService;
}

// Создаем и настраиваем контейнер
export const container = createContainer<Cradle>({
	injectionMode: InjectionMode.PROXY,
	strict: true,
});

// Регистрируем зависимости
container.register({
	// Bot как функция-фабрика
	bot: asFunction(() => {
		return createBot();
	}).singleton(),

	// Сервисы как классы
	invoiceService: asClass(InvoiceService).singleton(),
	notificationService: asClass(NotificationService).singleton(),
	duelService: asClass(DuelService).singleton(),
});

// Хелпер для получения контейнера с типизацией
export function getContainer(): AwilixContainer<Cradle> {
	return container;
}
