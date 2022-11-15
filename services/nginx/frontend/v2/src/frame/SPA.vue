<template>
  <div class="root">
    <header>
      <slot name="header"/>
    </header>
    <nav>
      <slot name="nav"/>
    </nav>
    <aside class="lsidebar">
      <slot name="lsidebar"/>
    </aside>
    <aside class="rsidebar">
      <slot name="rsidebar"/>
    </aside>
    <main>
      <slot/>
    </main>
    <footer>
      <slot name="footer"/>
    </footer>
  </div>
</template>

<script>

export default {
  name: 'SPA'
}
</script>

<style>
  .root {
    min-height: 100%;
    font-family: "Helvetica Neue",Helventica,Arial,sans-serif;
    font-size: 14px;
    line-height: 1.42857143;
    color: #333;
  }

  button {
    background-color: red;
    color: white;
  }
  button:disabled {
    opacity: 0.25;
  }
</style>
