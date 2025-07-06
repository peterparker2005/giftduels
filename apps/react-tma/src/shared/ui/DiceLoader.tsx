import { FC, useCallback, useEffect, useState } from "react";
import dice1 from "@/assets/lottie/dice/1.json";
import dice2 from "@/assets/lottie/dice/2.json";
import dice3 from "@/assets/lottie/dice/3.json";
import dice4 from "@/assets/lottie/dice/4.json";
import dice5 from "@/assets/lottie/dice/5.json";
import dice6 from "@/assets/lottie/dice/6.json";
import { useLottie } from "../hooks/useLottie";
import { LottiePlayer } from "./LottiePlayer";

const dices = [dice1, dice2, dice3, dice4, dice5, dice6];

function randomDice() {
	return dices[Math.floor(Math.random() * dices.length)];
}

export const DiceLoader: FC = () => {
	const [diceSrc, setDiceSrc] = useState(() => randomDice());
	const { lottie, lottieRefCallback } = useLottie();

	/** Перезапускаем новый JSON после завершения предыдущего */
	const handleComplete = useCallback(() => {
		setDiceSrc(randomDice());
	}, []);

	useEffect(() => {
		if (!lottie) return;

		lottie.setLoop(false);

		lottie.addEventListener("complete", handleComplete);
		lottie.play();

		return () => {
			lottie.removeEventListener("complete", handleComplete);
		};
	}, [lottie, handleComplete]);

	return (
		<div className="flex h-screen items-center justify-center">
			<LottiePlayer
				key={String(diceSrc)}
				className="size-40"
				src={diceSrc}
				autoplay
				loop={false}
				dotLottieRefCallback={lottieRefCallback}
			/>
		</div>
	);
};
