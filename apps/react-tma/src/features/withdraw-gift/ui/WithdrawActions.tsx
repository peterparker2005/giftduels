import { ExecuteWithdrawRequest_CommissionCurrency } from "@giftduels/protobuf-js/giftduels/gift/v1/gift_public_service_pb";
import { PreviewWithdrawResponse } from "@giftduels/protobuf-js/giftduels/payment/v1/public_service_pb";
import { useQueryClient } from "@tanstack/react-query";
import { openTelegramLink } from "@telegram-apps/sdk";
import { toast } from "sonner";
import { useExecuteWithdrawMutation } from "@/shared/api/queries/useExecuteWithdrawMutation";
import { logger } from "@/shared/logger";
import { Button } from "@/shared/ui/Button";
import { Icon } from "@/shared/ui/Icon/Icon";
import { formatTonAmount } from "@/shared/utils/formatTonAmount";

interface WithdrawActionsProps {
	giftIds: string[];
	previewData?: PreviewWithdrawResponse;
	onSuccess?: () => void;
	disabled?: boolean;
}

export const WithdrawActions = ({
	giftIds,
	previewData,
	onSuccess,
	disabled = false,
}: WithdrawActionsProps) => {
	const queryClient = useQueryClient();
	const { mutate: executeWithdraw, isPending: isExecuting } =
		useExecuteWithdrawMutation();

	const handleExecuteWithdraw = (
		commissionCurrency: ExecuteWithdrawRequest_CommissionCurrency,
	) => {
		if (giftIds.length === 0) {
			toast.error("No gifts selected", { position: "top-center" });
			return;
		}

		executeWithdraw(
			{
				giftIds,
				commissionCurrency,
			},
			{
				onSuccess: (data) => {
					console.log("Withdrawal successful");

					if (
						commissionCurrency === ExecuteWithdrawRequest_CommissionCurrency.TON
					) {
						queryClient.invalidateQueries({ queryKey: ["gifts"] });
						queryClient.invalidateQueries({ queryKey: ["balance"] });
						toast.success("Withdrawal successful!", { position: "top-center" });
					}

					if (
						commissionCurrency ===
						ExecuteWithdrawRequest_CommissionCurrency.STARS
					) {
						if (data.response.case === "starsInvoiceUrl") {
							const url = data.response.value;
							logger.debug("stars invoice url", url);

							// TODO: openInvoice instead of openTelegramLink
							openTelegramLink(url);
						} else {
							toast.error("Unexpected response from server", {
								position: "top-center",
							});
						}
						// обновляем баланс подарков
						queryClient.invalidateQueries({ queryKey: ["gifts"] });
					}

					// Call success callback
					onSuccess?.();
				},
				onError: (error) => {
					console.error("Withdrawal failed", error);
					toast.error(error.message || "Withdrawal failed", {
						position: "top-center",
					});
				},
			},
		);
	};

	return (
		<div className="space-y-2 mb-[calc(var(--tg-viewport-safe-area-inset-bottom)+16px)]">
			<Button
				onClick={() =>
					handleExecuteWithdraw(ExecuteWithdrawRequest_CommissionCurrency.TON)
				}
				disabled={disabled || isExecuting || giftIds.length === 0}
				className="w-full py-3 flex items-center justify-center gap-2 font-semibold"
				variant="primary"
			>
				{isExecuting ? (
					"Processing..."
				) : (
					<>
						<span>Withdraw for</span>
						<div className="flex items-center gap-2">
							<Icon icon="TON" className="w-5 h-5 shrink-0" />
							<span>
								{previewData?.totalTonFee?.value
									? formatTonAmount(previewData.totalTonFee.value)
									: "0.1"}
							</span>
						</div>
					</>
				)}
			</Button>
			<Button
				onClick={() =>
					handleExecuteWithdraw(ExecuteWithdrawRequest_CommissionCurrency.STARS)
				}
				disabled={disabled || isExecuting || giftIds.length === 0}
				className="w-full py-3 flex items-center justify-center gap-2 font-semibold bg-card-accent"
				variant="secondary"
			>
				{isExecuting ? (
					"Processing..."
				) : (
					<>
						<span>Withdraw for</span>
						<div className="flex items-center gap-2">
							<Icon icon="Star" className="w-5 h-5 shrink-0" />
							<span>{previewData?.totalStarsFee?.value || "10"}</span>
						</div>
					</>
				)}
			</Button>
		</div>
	);
};
