import { timestampDate } from "@bufbuild/protobuf/wkt";
import { TransactionReason } from "@giftduels/protobuf-js/giftduels/payment/v1/payment_pb";
import { useTransactionHistory } from "@/shared/api/queries/useTransactionHistory";
import { Button } from "@/shared/ui/Button";
import { cn } from "@/shared/utils/cn";
import { getFragmentUrl } from "@/shared/utils/getFragmentUrl";

export function TonTransactionHistory() {
	const { data, fetchNextPage } = useTransactionHistory();

	return (
		<section>
			{data?.pages.map((page) => (
				<div key={page.transactions.length} className="flex flex-col gap-2.5">
					{page.transactions.map((transaction) => (
						<div
							key={transaction.transactionId?.value}
							className="flex items-center justify-between"
						>
							<div className="flex items-center gap-4">
								<div className="w-14 h-14 rounded-2xl overflow-hidden bg-card">
									{transaction.metadata?.data.value?.slug && (
										<img
											// src={`https://nft.fragment.com/gift/${transaction.metadata?.data.value?.slug.toLowerCase()}.small.jpg`}
											src={getFragmentUrl(
												transaction.metadata?.data.value?.slug,
												"small",
											)}
											alt={transaction.metadata?.data.value?.slug}
											className="w-full h-full object-cover"
										/>
									)}
								</div>
								<div className="flex flex-col">
									<span className="font-medium">Withdraw Commission</span>
									{transaction.createdAt && (
										<span className="text-muted-foreground">
											{timestampDate(
												transaction.createdAt,
											).toLocaleTimeString()}
										</span>
									)}
								</div>
							</div>

							<span
								className={cn(
									"font-medium",
									transaction.reason === TransactionReason.WITHDRAW &&
										"text-foreground",
									transaction.reason === TransactionReason.DEPOSIT &&
										"text-success",
								)}
							>
								{transaction.tonAmount?.value}
							</span>
						</div>
					))}
				</div>
			))}
			<div>
				<Button onClick={() => fetchNextPage()}>Load more</Button>
			</div>
		</section>
	);
}
