export function getAvatarSlots(
	participants: Array<{ photoUrl?: string }>,
	maxPlayers: number,
): Array<string | null> {
	const slots: Array<string | null> = Array(maxPlayers).fill(null);

	participants.forEach((p, idx) => {
		if (idx < maxPlayers) {
			slots[idx] = p.photoUrl ?? null;
		}
	});

	return slots;
}
