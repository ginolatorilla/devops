<script lang="ts">
  let {
    toolName,
    toolDocumentation,
    headerHeight,
    footerHeight,
  }: {
    toolName: string;
    toolDocumentation: string[];
    headerHeight: number;
    footerHeight: number;
  } = $props();
</script>

<main
  class="flex flex-col px-4 pt-2
  md:pt-{headerHeight.toString()} md:flex-row md:gap-x-4 md:h-screen md:pb-{footerHeight.toString()}"
>
  <section
    class="order-first overflow-x-hidden
    md:order-last md:h-fit md:w-full md:overflow-y-auto"
  >
    <article class="p-2 my-2 border rounded-md md:my-4 border-amber-700">
      <p>Tool UI</p>
    </article>
  </section>

  <aside
    class="order-none overflow-x-hidden
    my-2 md:my-0 md:py-2 md:w-1/4 md:border-amber-700 md:border-r-1 md:overflow-y-auto"
  >
    <h1 class="my-1 text-xl font-bold">{toolName}</h1>
    {#each toolDocumentation as para}
      <p>{para}</p>
    {/each}
  </aside>
</main>
