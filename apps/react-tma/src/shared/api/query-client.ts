import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			// Время, в течение которого данные считаются свежими
			staleTime: 5 * 60 * 1000, // 5 минут

			// Время жизни кэша
			gcTime: 10 * 60 * 1000, // 10 минут (ранее называлось cacheTime)

			// Не перезапрашивать при фокусе окна
			refetchOnWindowFocus: false,

			// Не перезапрашивать при переподключении
			refetchOnReconnect: false,

			// Количество повторных попыток при ошибке
			retry: 2,

			// Задержка между повторными попытками
			retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
		},
		mutations: {
			// Количество повторных попыток для мутаций
			retry: 1,
		},
	},
});
