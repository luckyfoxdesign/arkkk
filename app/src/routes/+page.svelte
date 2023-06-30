<script>
	let result = ""
	async function sendRequestToCalculate(row) {
		const response = await fetch(`/calc?r=${row}`);
		const r = await response.json()
		result = r.r
	}
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Svelte demo app" />
</svelte:head>
<div class="wrapper">
	<div class="editor-input">
		<input class="input-line" placeholder="Ex: 1kg to gr" type="text" on:input={async (e) => await sendRequestToCalculate(e.target.value)}>
		<div class="input-result">= {result}</div>
	</div>
</div>

<style>
	.wrapper {
		display: flex;
		justify-content: center;
	}
	.editor-input {
		display: grid;
		grid-template-columns: 1fr;
		min-width: 100px;
		row-gap: 16px;
		font-family: Consolas, monospace;
	}
	.input-line {
		padding: 6px 8px;
		font-size: 19px;
	}
	.input-result {
		font-size: 24px;
	}
</style>
