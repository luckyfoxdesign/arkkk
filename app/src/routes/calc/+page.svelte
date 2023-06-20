<script>
let a = "",b = "", c = "", d = "";
	async function sendReq() {
		const response = await fetch(
			`/calc?r=${encodeURIComponent("1x to y")}`
		);
		const result = await response.json();
		a = result.fid
		b = result.tid
		c = result.val
		d = result.rqm
	}
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Svelte demo app" />
</svelte:head>

<button on:click={sendReq}>Send test req</button>
<p>{a}</p>
<p>{b}</p>
<p>{c}</p>
<p>{d}</p>