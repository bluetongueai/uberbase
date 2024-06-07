<script>
  import { Label, Input, Button, Spinner } from 'flowbite-svelte';
  import { EnvelopeSolid, LockSolid } from 'flowbite-svelte-icons';

  let className;
	export { className as class };
</script>

<form method="POST" action="/login" class="{className}">
  <div class="my-3">
    <Input id="email" type="email" placeholder="name@flowbite.com">
      <EnvelopeSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
    </Input>
  </div>
  <div class="my-3">
    <Input id="password" type="password" placeholder="password">
      <LockSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
    </Input>
  </div>
  <div class="text-right">
    <Button outline>
      <Spinner class="me-3" size="3" />
      Login
    </Button>
  </div>
</form>
