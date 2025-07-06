export function encoderFor<T>(messageModule: {
	encode: (m: T) => any
	$type: string
}) {
	return {
		encode: messageModule.encode,
		$type: messageModule.$type,
	}
}
