import { useNavigate } from "@tanstack/react-router";
import { useBalanceQuery } from "@/shared/api/queries/useBalanceQuery";
import { Icon } from "@/shared/ui/Icon/Icon";
import { formatTonAmount } from "@/shared/utils/formatTonAmount";

export const TonBalance = () => {
	const { data, isLoading } = useBalanceQuery();
	const navigate = useNavigate();
	if (isLoading) return <div>Loading...</div>;

	return (
		<button
			onClick={() => {
				navigate({ to: "/balance" });
			}}
			type="button"
			className="flex items-center gap-2 bg-[#2D9EED]/15 rounded-full px-2 py-1.5 blur-background"
		>
			<div className="w-5 h-5 flex items-center justify-center bg-[#2D9EED] rounded-full">
				<Icon icon="TON" className="w-3.5 h-3.5" />
			</div>
			<span className="font-medium">
				{formatTonAmount(data?.balance?.tonAmount?.value)} TON
			</span>
			<div className="w-4 h-4 flex items-center justify-center bg-white rounded-full">
				<Icon icon="Plus" className="w-3.5 h-3.5 text-background" />
			</div>
		</button>
	);
};
