<script lang="ts">
  let {
    copyright,
    height,
  }: {
    copyright: string;
    height: number;
  } = $props();
</script>

<footer
  class="px-4 py-2 mt-5 md:h-{height.toString()} md:py-1 md:fixed md:bottom-0 md:flex md:flex-row md:justify-end md:w-full md:bg-amber-700 md:text-amber-50 text-stone-400"
>
  <p class="text-sm md:text-xs">{copyright}</p>
</footer>
