import { create } from "@bufbuild/protobuf";
import { GiftIdSchema } from "@giftduels/protobuf-js/giftduels/shared/v1/common_pb";
import { useQuery } from "@tanstack/react-query";
import { paymentClient } from "@/shared/api/client";

const Page = () => {
	const { data } = useQuery({
		queryKey: ["testing"],
		queryFn: () =>
			paymentClient.previewWithdraw({
				giftIds: [
					create(GiftIdSchema, {
						value: "e5551d57-1920-4106-aa10-ce0eb22bce83",
					}),
				],
			}),
	});
	return (
		<div className="container">
			{data?.fees?.map((fee) => (
				<div key={fee.giftId?.value}>
					<h1>{fee.tonFee?.value}</h1>
					<h1>{fee.starsFee?.value}</h1>
				</div>
			))}
		</div>
	);
};

export default Page;
