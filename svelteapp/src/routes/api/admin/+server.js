import { json } from "@sveltejs/kit";

export async function PUT({ event, url, params, request }) {
  const reqType = url.searchParams.get("t");
  const reqBodyData = await request.text()
  const textWithSpaces = reqBodyData.replaceAll(/\n/g, " ")
  let insertStatus = 0;

  if (reqType === "w") {
    insertStatus = await getResult(textWithSpaces);
  }

  return json({ status: insertStatus });
}

async function getResult(text) {
  const resp = await fetch(`http://sc_goapp:8080/insertkv`, {
    method: "PUT",
    headers: { "Content-Type": "text/plain" },
    body: text,
  })
  console.log(resp.status);
  console.log(resp.statusText);
  return 200;
}
