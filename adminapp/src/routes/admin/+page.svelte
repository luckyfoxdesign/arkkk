<script>
  let textarea = "";
  let status = "";
  async function sendData() {
    const text = textarea.replaceAll("\n", " ");
    const response = await fetch(`/api/admin/?t=${encodeURIComponent("w")}`, {
      method: "PUT",
      headers: { "Content-Type": "text/plain" },
      body: text,
    });
    const responseJSON = await response.json();
	if (responseJSON.status == 200) {
		textarea = ""
		status = "inserted"
		// TODO
		// goapp sends several different statues
		// need to handle them all
		setTimeout(() => {
			status = ""
		}, 3000)
	}
  }

  // TODO
  // to add:
  // update key
  // remove key
  // get keys by id

  export let data;
</script>

<svelte:head>
  <title>Home</title>
</svelte:head>

<textarea name="" id="" cols="30" rows="10" bind:value={textarea} />
<p>Status: {status}</p>
<button type="submit" on:click={sendData}>Add to DB</button>

<div class="list">
  {#each data.list as pair}
    <div class="list-pair">
      <label>{pair.Val}<input type="text" value="{pair.Key}"></label>
      <button type="submit">Save</button>
    </div>
  {/each}
</div>
