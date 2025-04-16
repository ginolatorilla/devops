<script lang="ts">
  let {
    title,
    tools,
    externalURL,
    height,
  }: {
    title: string;
    tools: string[];
    externalURL: string;
    height: number;
  } = $props();
</script>

<header
  class="sticky top-0 flex flex-row items-center justify-between w-full h-{height.toString()} px-4 py-2
bg-amber-700 text-cyan-50
  md:fixed"
>
  <h1 class="text-lg font-bold">{title}</h1>
  <nav class="flex flex-row gap-x-5">
    {#each tools as tool}
      <p>{tool}</p>
    {/each}
  </nav>
  <a
    class="underline"
    href={externalURL}
    target="_blank"
    rel="noopener noreferrer"
  >
    Visit GitHub
  </a>
</header>
