export async function load({ params }) {
  const r = await getAllKeyValuePairs()
  return {
    list: r.list,
  };
}

async function getAllKeyValuePairs() {
  const requestResponse = await fetch(`http://sc_goapp:8080/keys/list`);
  const responseJSON = await requestResponse.json();
  return responseJSON;
}
