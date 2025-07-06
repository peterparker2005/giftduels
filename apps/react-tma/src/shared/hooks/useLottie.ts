import { DotLottie } from "@lottiefiles/dotlottie-react";
import React from "react";

/**
 * Хук для управления анимацией DotLottie
 * Позволяет получать экземпляр DotLottie из компонента DotLottieReact.
 */
export function useLottie() {
	const [lottie, setLottie] = React.useState<DotLottie | null>(null);

	/**
	 * Колбэк для получения инстанса DotLottie из компонента DotLottieReact.
	 * @param {DotLottie} dotLottie - экземпляр DotLottie
	 */
	const lottieRefCallback = (lottie: DotLottie) => {
		setLottie(lottie);
	};

	return { lottie, lottieRefCallback };
}
