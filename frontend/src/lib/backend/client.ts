import type Example from "$lib/components/Example.svelte";
import type { ExampleSet } from "$lib/models/ExampleData";
import type { Input } from "$lib/models/Input";
import type { ParsedSentence } from "$lib/models/ParsedSentence";

export class BackendClient {
  private fetcher: typeof fetch;

  constructor(fetcher: typeof fetch) {
    this.fetcher = fetcher;
  }

  public getExamples(query: string): Promise<ExampleSet[]> {
    return this.fetch<{examples: ExampleSet[]}>(
      `/api/examples/${encodeURIComponent(query)}`
    ).then(j => j.examples);
  }

  public postExample(input: Input, dry = false): Promise<Example> {
    return this.fetch<{example: Example}>(
      `/api/examples/?dry=${!!dry?"true":"false"}`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(input)
      }
    ).then(j => j.example);
  }

  public deleteExample(id: string): Promise<Example> {
    return this.fetch<{example: Example}>(
      `/api/examples/${id}`, {method: "DELETE"}
    ).then(j => j.example);
  }

  public getInput(id: string): Promise<Input> {
    return this.fetch<{input: Input}>(
      `/api/examples/${id}/input`
    ).then(j => j.input);
  }

  public parseSentence(text: string, lookup: boolean = true): Promise<ParsedSentence> {
    return this.fetch<ParsedSentence>(
      `/api/utils/parse-sentence`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({text, lookup})
      }
    );
  }


  private async fetch<T>(path: string, init?: RequestInit): Promise<T> {
    const url = (import.meta.env.VITE_BACKEND_URL||"") + path;
  
    const res = await this.fetcher(url, init);
    if (res.status !== 200) {
      if (res.headers.get("Content-Type")?.includes("application/json")) {
        throw await res.json();
      } else {
        throw await res.text();
      }
    }
  
    const json = await res.json();
  
    return json as T;
  }
}