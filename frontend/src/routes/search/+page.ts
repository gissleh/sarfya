import { BackendClient } from "$lib/backend/client";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async({ url, fetch }) => {
	try {
		const exampleSets = await (new BackendClient(fetch)).getExamples(url.searchParams.get("q")||"")
		return { exampleSets, query: url.searchParams.get("q"), error: "" };
	} catch(err: any) {
		return { exampleSets: [], query: url.searchParams.get("q"), error: err.error || err?.toString() || err };
	}
}

