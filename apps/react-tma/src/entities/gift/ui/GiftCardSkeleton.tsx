export const GiftCardSkeleton = () => {
	return (
		<div className="bg-card rounded-lg p-4 animate-pulse">
			<div className="aspect-square bg-muted rounded-lg mb-3"></div>
			<div className="space-y-2">
				<div className="h-4 bg-muted rounded w-3/4"></div>
				<div className="h-3 bg-muted rounded w-1/2"></div>
			</div>
		</div>
	);
};
