export async function load({ params }) {
  const r = await getAllKeyValuePairs()
  return {
    list: r.list,
  };
}

async function getAllKeyValuePairs() {
  let responseJSON = []
  const requestResponse = await fetch(`http://sc_goapp:8080/keys/list`);

  if (requestResponse.status == 200) {
    responseJSON = await requestResponse.json();
  }
  return responseJSON;
}