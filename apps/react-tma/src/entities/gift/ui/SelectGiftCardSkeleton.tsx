export function SelectGiftCardSkeleton() {
	return (
		<div className="bg-card rounded-lg p-3 animate-pulse">
			<div className="flex items-center gap-3">
				<div className="w-12 h-12 bg-muted rounded-lg flex-shrink-0"></div>
				<div className="flex-1 space-y-2">
					<div className="h-4 bg-muted rounded w-3/4"></div>
					<div className="h-3 bg-muted rounded w-1/2"></div>
				</div>
				<div className="w-5 h-5 bg-muted rounded-full flex-shrink-0"></div>
			</div>
		</div>
	);
}
