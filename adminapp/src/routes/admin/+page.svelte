<script>
  export let data;
  const tempData = JSON.parse(JSON.stringify(data));

  let textarea = "";
  let status = "";
  let sortByNameInput = "";
  let sortByIdInput = "";
  async function sendData() {
    const text = textarea.replaceAll("\n", " ");
    const response = await fetch(`/api/?t=${encodeURIComponent("w")}`, {
      method: "PUT",
      headers: { "Content-Type": "text/plain" },
      body: text,
    });
    const responseJSON = await response.json();
    if (responseJSON.status == 200) {
      textarea = "";
      status = "inserted";
      // TODO
      // goapp sends several different statues
      // need to handle them all
      setTimeout(() => {
        status = "";
      }, 3000);
    }
  }

  // TODO
  // to add:
  // update key
  // remove key
  // get keys by id

  function sortById() {
    let arr = [];
    const a = [...tempData.list];

    a.map((e) => {
      if (e.Val.includes(sortByIdInput)) {
        let o = {
          Key: e.Key,
          Val: e.Val,
        };
        arr.push(o);
      }
    });

    if (arr.length != 0) {
      data.list = arr;
    }
    if (sortByIdInput.length == 0) {
      data.list = a;
    }
  }

  function sortByName() {
    let arr = [];
    const a = [...tempData.list];

    a.map((e) => {
      if (e.Key.includes(sortByNameInput)) {
        let o = {
          Key: e.Key,
          Val: e.Val,
        };
        arr.push(o);
      }
    });

    if (arr.length != 0) {
      data.list = arr;
    }
    if (sortByNameInput.length == 0) {
      data.list = a;
    }
  }
</script>

<svelte:head>
  <title>Home</title>
</svelte:head>

<div class="content">
  <div class="insert-field">
    <textarea name="" id="" cols="30" rows="10" bind:value={textarea} />
    <p>Status: {status}</p>
    <button type="submit" on:click={sendData}>Add to DB</button>
  </div>

  <div class="units-list">
    <h3>Count: {data.list.length}</h3>
    <div class="list">
      <div class="filters">
        <label>
          Sort by id:
          <input type="text" bind:value={sortByIdInput} on:input={sortById} />
        </label>
        <label>
          Sort by name:
          <input type="text" bind:value={sortByNameInput} on:input={sortByName} />
        </label>
      </div>
      <div class="list-values">
        {#each data.list as pair}
          <div class="list-pair">
            <label>{pair.Val}<input type="text" value={pair.Key} /></label>
            <button type="submit">Save</button>
            <button type="submit">Remove</button>
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>

<style>
  .content {
    display: grid;
    grid-template-columns: auto max-content 1fr;
    grid-column-gap: 32px;
  }
  .filters {
    top: 0;
    position: sticky;
    background-color: white;
    padding-top: 16px;
    padding-bottom: 16px;
  }
  .list {
    padding-left: 16px;
    padding-right: 16px;
    border: 1px solid #cccccc;
    height: 400px;
    overflow-y: scroll;
  }
  .list-values {
    display: grid;
    grid-row-gap: 16px;
  }
</style>
