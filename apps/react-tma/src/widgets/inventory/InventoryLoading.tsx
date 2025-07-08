export const InventoryLoading = () => {
	return (
		<section className="grid grid-cols-2 gap-4">
			{Array.from({ length: 4 }).map((_, index) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: skeleton key
				<div key={index} className="p-2.5 rounded-3xl bg-card animate-pulse">
					<div className="relative w-full h-40 rounded-2xl cursor-default" />
					<div className="mt-2 h-9" />
				</div>
			))}
		</section>
	);
};
