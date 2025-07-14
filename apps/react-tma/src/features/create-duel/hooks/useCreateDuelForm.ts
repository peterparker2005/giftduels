import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

const createDuelSchema = z.object({
	privacy: z.enum(["public", "private"]),
	players: z.number().min(2).max(4),
	gifts: z.array(z.string().min(1)).min(1, "At least one gift is required"),
});

type CreateDuelForm = z.infer<typeof createDuelSchema>;

export function useCreateDuelForm() {
	const form = useForm<CreateDuelForm>({
		resolver: zodResolver(createDuelSchema),
		defaultValues: {
			privacy: "public",
			players: 2,
			gifts: [],
		},
	});

	return form;
}
