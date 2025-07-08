import { cn } from "@/shared/utils/cn";
import { useProfileQuery } from "../api/queries/useProfileQuery";

interface AvatarProps extends React.HTMLAttributes<HTMLImageElement> {
	className?: string;
}

export function Avatar({ className, ...props }: AvatarProps) {
	const { data } = useProfileQuery();

	return (
		<img
			src={data?.profile?.photoUrl}
			alt="avatar"
			className={cn("w-10 h-10 rounded-full", className)}
			{...props}
		/>
	);
}
