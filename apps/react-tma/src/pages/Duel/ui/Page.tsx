import { useParams } from "@tanstack/react-router";
import { JoinDuelDrawer } from "@/features/join-duel";
import { useDuelQuery } from "@/shared/api/queries/useDuelQuery";
import { Button } from "@/shared/ui/Button";

export function Page() {
	const { duelId } = useParams({ from: "/default-layout/duel/$duelId" });
	const { data, isLoading } = useDuelQuery({ duelId });

	if (isLoading) return <div>Loading...</div>;

	const handleJoinDuel = async (data: any) => {
		console.log("Joining duel with data:", data);
		// TODO: Implement actual join duel API call
		// const response = await joinDuelMutation.mutateAsync({
		//   duelId,
		//   selectedGifts: data.selectedGifts,
		//   entryPrice: data.entryPrice,
		// });
	};

	return (
		<div className="container">
			<div className="flex flex-col gap-4">
				<div className="text-center">
					<h1 className="text-2xl font-bold">Duel #{duelId}</h1>
					{data?.duel && (
						<div className="mt-2 text-muted-foreground">
							<p>Status: {data.duel.status}</p>
							<p>
								Players: {data.duel.participants?.length || 0}/
								{data.duel.params?.maxPlayers || 0}
							</p>
						</div>
					)}
				</div>

				{data?.duel && (
					<JoinDuelDrawer
						displayNumber={data.duel.displayNumber.toString()}
						duel={data.duel}
						onJoinDuel={handleJoinDuel}
					>
						<Button className="w-full py-3">Join Duel</Button>
					</JoinDuelDrawer>
				)}

				{/* Duel Details */}
				{data?.duel && (
					<div className="space-y-4">
						<div className="bg-card rounded-lg p-4">
							<h3 className="font-semibold mb-2">Duel Information</h3>
							<div className="space-y-1 text-sm">
								<p>Max Players: {data.duel.params?.maxPlayers}</p>
								<p>Max Gifts: {data.duel.params?.maxGifts}</p>
								<p>Private: {data.duel.params?.isPrivate ? "Yes" : "No"}</p>
								<p>
									Total Stake Value: {data.duel.totalStakeValue?.value || "0"}{" "}
									TON
								</p>
							</div>
						</div>

						{/* Participants */}
						{data.duel.participants && data.duel.participants.length > 0 && (
							<div className="bg-card rounded-lg p-4">
								<h3 className="font-semibold mb-2">Participants</h3>
								<div className="space-y-2">
									{data.duel.participants.map((participant) => (
										<div
											key={
												participant.telegramUserId?.value ||
												`participant-${Math.random()}`
											}
											className="flex items-center gap-2"
										>
											{participant.photoUrl && (
												<img
													src={participant.photoUrl}
													alt="Participant"
													className="w-8 h-8 rounded-full"
												/>
											)}
											<span className="text-sm">
												{participant.telegramUserId?.value}
												{participant.isCreator && " (Creator)"}
											</span>
										</div>
									))}
								</div>
							</div>
						)}

						{/* Stakes */}
						{data.duel.stakes && data.duel.stakes.length > 0 && (
							<div className="bg-card rounded-lg p-4">
								<h3 className="font-semibold mb-2">Stakes</h3>
								<div className="space-y-2">
									{data.duel.stakes.map((stake) => (
										<div
											key={
												stake.gift?.giftId?.value || `stake-${Math.random()}`
											}
											className="flex items-center justify-between"
										>
											<span className="text-sm">{stake.gift?.title}</span>
											<span className="text-sm font-medium">
												{stake.stakeValue?.value || "0"} TON
											</span>
										</div>
									))}
								</div>
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
}

export default Page;
