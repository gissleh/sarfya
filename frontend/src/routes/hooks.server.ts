import type { Handle } from "@sveltejs/kit";

export const handle: Handle = ({ event, resolve }) => {
	return resolve(event, {
		// you can also seralize other headers here if needed
		filterSerializedResponseHeaders: (name, value) => true
	});
}