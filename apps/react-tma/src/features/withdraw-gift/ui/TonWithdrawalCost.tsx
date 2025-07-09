import { AnimatePresence, motion } from "motion/react";
import { useEffect, useState } from "react";
import { BiLoader } from "react-icons/bi";

interface TonWithdrawalCostProps {
	isPending: boolean;
	fee: number | undefined;
}

export function TonWithdrawalCost({ isPending, fee }: TonWithdrawalCostProps) {
	const [showLoader, setShowLoader] = useState(false);

	useEffect(() => {
		if (isPending) {
			setShowLoader(true);
		} else {
			const timer = setTimeout(() => setShowLoader(false), 500);
			return () => clearTimeout(timer);
		}
	}, [isPending]);

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
					{fee ?? "0.15"}
				</motion.p>
			)}
		</AnimatePresence>
	);
}
