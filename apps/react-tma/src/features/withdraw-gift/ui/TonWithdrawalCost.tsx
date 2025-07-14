import { AnimatePresence, motion } from "motion/react";
import { useEffect, useRef, useState } from "react";
import { BiLoader } from "react-icons/bi";

interface TonWithdrawalCostProps {
	isPending: boolean;
	fee: number | undefined;
}

export function TonWithdrawalCost({ isPending, fee }: TonWithdrawalCostProps) {
	const [showLoader, setShowLoader] = useState(false);
	const lastFeeRef = useRef<number>(fee ?? 0.1); // стартовое значение

	useEffect(() => {
		if (fee !== undefined) {
			lastFeeRef.current = fee;
		}
	}, [fee]);

	useEffect(() => {
		if (isPending) {
			setShowLoader(true);
		} else {
			const timer = setTimeout(() => setShowLoader(false), 500);
			return () => clearTimeout(timer);
		}
	}, [isPending]);

	const displayFee = lastFeeRef.current;

	return (
		<AnimatePresence mode="wait" initial={false}>
			{showLoader ? (
				<motion.div
					key="loader"
					initial={{ opacity: 0 }}
					animate={{ opacity: 1 }}
					exit={{ opacity: 0 }}
					transition={{ duration: 0.2 }}
				>
					<BiLoader className="animate-spin" />
				</motion.div>
			) : (
				<motion.p
					key="fee"
					initial={{ opacity: 0 }}
					animate={{ opacity: 1 }}
					exit={{ opacity: 0 }}
					transition={{ duration: 0.2 }}
				>
					{displayFee}
				</motion.p>
			)}
		</AnimatePresence>
	);
}
