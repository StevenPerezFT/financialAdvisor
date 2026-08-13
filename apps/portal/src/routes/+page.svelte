<script>
  const modules = [
    { title: 'Finance', href: '/finance' },
    { title: 'Movements', href: '/finance/movements' },
    { title: 'Limits', href: '/finance/limits' }
  ];
</script>

<svelte:head>
  <title>Financial Advisor</title>
</svelte:head>

<main style="padding: 2rem; font-family: sans-serif;">
  <h1>Financial Advisor</h1>
  <p>Starting panel to manage finances and tax advisory.</p>

  <ul>
    {#each modules as module}
      <li><a href={module.href}>{module.title}</a></li>
    {/each}
  </ul>
</main>
